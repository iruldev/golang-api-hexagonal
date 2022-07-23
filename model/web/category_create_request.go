package web

type CategoryCreateRequest struct {
	Name string `validate:"required,max=225,min=1" json:"name"`
}
