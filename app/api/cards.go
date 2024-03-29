package api

import (
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"

	"github.com/baza-trainee/ataka-help-backend/app/logger"
	"github.com/baza-trainee/ataka-help-backend/app/structs"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type CardService interface {
	ReturnCards(context.Context, structs.CardQueryParameters) ([]structs.Card, int, error)
	SaveCard(context.Context, *multipart.Form) error
	ReturnCardByID(context.Context, string) (structs.Card, error)
	DeleteCardByID(context.Context, string) error
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
		Limit: defaultLimit,
		Page:  defaultPage,
	}

	if err := ctx.QueryParser(&params); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if params.Limit < 0 || params.Page < 0 {
		return fiber.NewError(fiber.StatusBadRequest, "query params values cant't be negative")
	}

	if params.Limit > 0 && params.Page == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "page values should be grater then 0")
	}

	cards, total, err := h.Service.ReturnCards(ctx.UserContext(), params)
	if err != nil && !errors.Is(err, structs.ErrNotFound) {
		return fiber.NewError(fiber.StatusNoContent, err.Error())
	}

	response := structs.CardsResponse{
		Code:  fiber.StatusOK,
		Total: total,
		Cards: cards,
	}

	return ctx.Status(fiber.StatusOK).JSON(response) // nolint
}

// nolint: cyclop
func (h CardHandler) createCard(ctx *fiber.Ctx) error {
	allowedFileExtentions := []string{"jpg", "jpeg", "webp", "png"}

	const (
		limitNumberItemsFile = 1
		minAltItems          = 10
		maxAltItems          = 30
		minTitle             = 3
		maxTitle             = 150
		minDescription       = 3
	)

	form, err := ctx.MultipartForm()
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if len(form.File["thumb"]) < limitNumberItemsFile {
		h.log.Debugw("createCard", "form.File", "no thumb was attached")

		return fiber.NewError(fiber.StatusBadRequest, "no thumb was attached")
	}

	fileHeader := form.File["thumb"][0]

	switch {
	case fileHeader == nil || fileHeader.Size > fileLimit || !isAllowedFileExtention(allowedFileExtentions, fileHeader.Filename):
		h.log.Debugw("createCard", "form.File", "required thumb not biger then 5 Mb and format jpg/jpeg/webp/png")

		return fiber.NewError(fiber.StatusBadRequest, "required thumb not bigger then 5 Mb and format jpg/jpeg/webp/png")
	case form.Value["title"] == nil || symbolsCounter(form.Value["title"][0]) < minTitle || symbolsCounter(form.Value["title"][0]) > maxTitle:
		h.log.Debugw("createCard", "form.Vlaues", "required title atleast 3 letters and less than 150")

		return fiber.NewError(fiber.StatusBadRequest, "required title atleast 3 letters and less than 150")
	case form.Value["alt"] == nil || symbolsCounter(form.Value["alt"][0]) < minAltItems || symbolsCounter(form.Value["alt"][0]) > maxAltItems:
		h.log.Debugw("createCard", "form.Vlaues", "alt is out of limit range")

		return fiber.NewError(fiber.StatusBadRequest, "alt is out of limit range")
	case form.Value["description"] == nil || len(form.Value["description"][0]) < minDescription:
		h.log.Debugw("createCard", "form.Vlaues", "required description")

		return fiber.NewError(fiber.StatusBadRequest, "required description")
	}

	descriptions := []string{}

	err = json.Unmarshal([]byte(form.Value["description"][0]), &descriptions)
	if err != nil {
		h.log.Debugw("Unmarshal", "description", err.Error())

		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if err := h.Service.SaveCard(ctx.Context(), form); err != nil {
		if errors.Is(err, structs.ErrUniqueRestriction) {
			h.log.Errorw("SaveCard", "error", err.Error())

			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return ctx.Status(fiber.StatusCreated).JSON(structs.SetResponse(fiber.StatusCreated, "success")) // nolint
}

func (h CardHandler) findCard(ctx *fiber.Ctx) error {
	param := struct {
		ID string `params:"id"`
	}{}

	if err := ctx.ParamsParser(&param); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	_, err := uuid.Parse(param.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "id is not uuid type")
	}

	card, err := h.Service.ReturnCardByID(ctx.Context(), param.ID)
	if err != nil {
		if errors.Is(err, structs.ErrNotFound) {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return ctx.Status(fiber.StatusOK).JSON(card)
}

func (h CardHandler) deleteCard(ctx *fiber.Ctx) error {
	param := struct {
		ID string `params:"id"`
	}{}

	if err := ctx.ParamsParser(&param); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	_, err := uuid.Parse(param.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "id is not uuid type")
	}

	if err := h.Service.DeleteCardByID(ctx.Context(), param.ID); err != nil {
		if errors.Is(err, structs.ErrNoRowAffected) {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return ctx.Status(fiber.StatusOK).JSON(structs.SetResponse(fiber.StatusOK, "success"))
}
