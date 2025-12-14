// Package note provides the gRPC handler for the NoteService.
// It follows hexagonal architecture by calling the usecase layer,
// not the infrastructure layer directly.
package note

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	notedom "github.com/iruldev/golang-api-hexagonal/internal/domain/note"
	noteuc "github.com/iruldev/golang-api-hexagonal/internal/usecase/note"
	notev1 "github.com/iruldev/golang-api-hexagonal/proto/note/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Handler implements the NoteServiceServer interface.
// It injects the usecase layer following hexagonal architecture.
type Handler struct {
	notev1.UnimplementedNoteServiceServer
	usecase *noteuc.Usecase
}

// NewHandler creates a new gRPC note handler.
func NewHandler(uc *noteuc.Usecase) *Handler {
	return &Handler{usecase: uc}
}

// CreateNote creates a new note.
// Returns INVALID_ARGUMENT if title is empty or too long.
func (h *Handler) CreateNote(ctx context.Context, req *notev1.CreateNoteRequest) (*notev1.CreateNoteResponse, error) {
	note, err := h.usecase.Create(ctx, req.GetTitle(), req.GetContent())
	if err != nil {
		return nil, mapErrorToStatus(err)
	}

	return &notev1.CreateNoteResponse{
		Note: toProto(note),
	}, nil
}

// GetNote retrieves a note by its ID.
// Returns NOT_FOUND if the note doesn't exist.
func (h *Handler) GetNote(ctx context.Context, req *notev1.GetNoteRequest) (*notev1.GetNoteResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid note id: %v", err)
	}

	note, err := h.usecase.Get(ctx, id)
	if err != nil {
		return nil, mapErrorToStatus(err)
	}

	return &notev1.GetNoteResponse{
		Note: toProto(note),
	}, nil
}

// ListNotes retrieves a paginated list of notes.
func (h *Handler) ListNotes(ctx context.Context, req *notev1.ListNotesRequest) (*notev1.ListNotesResponse, error) {
	pageSize := int(req.GetPageSize())
	if pageSize <= 0 {
		pageSize = 10 // default page size
	}
	if pageSize > 100 {
		pageSize = 100 // max page size
	}

	// Decode page token to get page number
	page := 1
	if req.GetPageToken() != "" {
		decoded, err := base64.StdEncoding.DecodeString(req.GetPageToken())
		if err == nil {
			if p, err := strconv.Atoi(string(decoded)); err == nil && p > 0 {
				page = p
			}
		}
	}

	notes, totalCount, err := h.usecase.List(ctx, page, pageSize)
	if err != nil {
		return nil, mapErrorToStatus(err)
	}

	// Convert domain notes to proto messages
	protoNotes := make([]*notev1.Note, 0, len(notes))
	for _, n := range notes {
		protoNotes = append(protoNotes, toProto(n))
	}

	// Generate next page token if there are more results
	var nextPageToken string
	if int64(page*pageSize) < totalCount {
		nextPageToken = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", page+1)))
	}

	return &notev1.ListNotesResponse{
		Notes:         protoNotes,
		NextPageToken: nextPageToken,
		TotalCount:    totalCount,
	}, nil
}

// UpdateNote updates an existing note.
// Returns NOT_FOUND if the note doesn't exist.
// Returns INVALID_ARGUMENT if title is empty or too long.
func (h *Handler) UpdateNote(ctx context.Context, req *notev1.UpdateNoteRequest) (*notev1.UpdateNoteResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid note id: %v", err)
	}

	note, err := h.usecase.Update(ctx, id, req.GetTitle(), req.GetContent())
	if err != nil {
		return nil, mapErrorToStatus(err)
	}

	return &notev1.UpdateNoteResponse{
		Note: toProto(note),
	}, nil
}

// DeleteNote removes a note by its ID.
// Returns NOT_FOUND if the note doesn't exist.
func (h *Handler) DeleteNote(ctx context.Context, req *notev1.DeleteNoteRequest) (*notev1.DeleteNoteResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid note id: %v", err)
	}

	if err := h.usecase.Delete(ctx, id); err != nil {
		return nil, mapErrorToStatus(err)
	}

	return &notev1.DeleteNoteResponse{}, nil
}

// toProto converts a domain Note to a proto Note message.
func toProto(n *notedom.Note) *notev1.Note {
	return &notev1.Note{
		Id:        n.ID.String(),
		Title:     n.Title,
		Content:   n.Content,
		CreatedAt: timestamppb.New(n.CreatedAt),
		UpdatedAt: timestamppb.New(n.UpdatedAt),
	}
}

// mapErrorToStatus maps domain errors to gRPC status codes.
func mapErrorToStatus(err error) error {
	switch {
	case errors.Is(err, notedom.ErrNoteNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, notedom.ErrEmptyTitle):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, notedom.ErrTitleTooLong):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
