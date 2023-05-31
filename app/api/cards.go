package api

import (
	"context"
	"errors"
	"mime/multipart"
	"net/http"

	"github.com/baza-trainee/ataka-help-backend/app/logger"
	"github.com/baza-trainee/ataka-help-backend/app/structs"
	"github.com/gofiber/fiber/v2"
)

const (
	fileLimit     = 5 * 1024 * 1024
	defaultLimit  = 6
	defaultOffset = 0
)

var allowedContentType = []string{"image/jpg", "image/jpeg", "image/webp"}

type CardService interface {
	ReturnCards(structs.CardQueryParameters, context.Context) ([]structs.Card, int, error)
	SaveCard(*multipart.Form, *fiber.Ctx) error
}

type CardHandler struct {
	Service CardService
	log     *logger.Logger
}

func NewCardsHandler(service CardService, log *logger.Logger) CardHandler {
	return CardHandler{
		Service: service,
		log:     log,
	}
}

func (h CardHandler) getCards(ctx *fiber.Ctx) error {
	params := structs.CardQueryParameters{
		Limit:  defaultLimit,
		Offset: defaultOffset,
	}

	if err := ctx.QueryParser(&params); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	cards, total, err := h.Service.ReturnCards(params, ctx.Context())
	if err != nil && !errors.Is(err, structs.ErrNotFound) {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	response := structs.CardsResponse{
		Code:  fiber.StatusOK,
		Tolal: total,
		Cards:  cards,
	}
	if err := ctx.JSON(response); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return nil
}

func (h CardHandler) createCard(ctx *fiber.Ctx) error {
	form, err := ctx.MultipartForm()
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	fileHeader := form.File["thumb"][0]

	if fileHeader == nil || fileHeader.Size > fileLimit || !isAllowedContentType(allowedContentType, fileHeader.Header["Content-Type"][0]) {
		h.log.Debugw("createCard", "form.File", "required thumb not biger then 5 Mb and format jpg/jpeg/webp")
		return fiber.NewError(fiber.StatusBadRequest, "required thumb not biger then 5 Mb and format jpg/jpeg/webp")
	}

	if form.Value["title"] == nil || len(form.Value["title"][0]) < 4 || len(form.Value["title"][0]) > 300 {
		h.log.Debugw("createCard", "form.Vlaues", "required title more than 3 letters and less than 300")
		return fiber.NewError(fiber.StatusBadRequest, "required title more than 3 letters and less than 300")
	}

	if form.Value["alt"] == nil || len(form.Value["alt"][0]) < 1 {
		h.log.Debugw("createCard", "form.Vlaues", "required alt")
		return fiber.NewError(fiber.StatusBadRequest, "required alt")
	}

	if form.Value["description"] == nil || len(form.Value["description"][0]) < 2 {
		h.log.Debugw("createCard", "form.Vlaues", "required description")
		return fiber.NewError(fiber.StatusBadRequest, "required description")
	}

	if err := h.Service.SaveCard(form, ctx); err != nil {
		if errors.Is(err, structs.ErrUniqueRestriction) {
			h.log.Errorw("SaveCard", "error", err.Error())
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	if err := ctx.JSON(structs.SetResponse(http.StatusOK, "success")); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return nil
}

func isAllowedContentType(allowedList []string, contentType string) bool {
	for _, i := range allowedList {
		if i == contentType {
			return true
		}
	}

	return false
}