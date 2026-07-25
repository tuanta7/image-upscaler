package image

type Handler struct {
	uc *UseCase
}

type UpscaleRequest struct {
	Image []byte
	Scale int
}
