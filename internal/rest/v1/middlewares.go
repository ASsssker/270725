package v1

import (
	bp "270725/internal/rest/v1/boileplate"
	"270725/internal/services"
	"errors"
	"github.com/labstack/echo/v4"
	"net/http"
)

func (h *Handler) handleError() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			err = next(c)
			if err != nil {
				h.log.Error(err.Error())

				switch {
				case errors.Is(err, services.ErrValidation):
					// TODO: Лучше завести кастомную ошибки, будет легче извлекать текст ошибки валидаци
					return c.JSON(http.StatusBadRequest, bp.Error{
						ErrorCode:   http.StatusBadRequest,
						Description: err.Error(),
					})

				case errors.Is(err, services.ErrTaskNotFound):
					return c.JSON(http.StatusNotFound, bp.Error{
						ErrorCode:   http.StatusNotFound,
						Description: "task not found",
					})

				case errors.Is(err, services.ErrServiceBusy):
					return c.JSON(http.StatusTooManyRequests, bp.Error{
						ErrorCode:   http.StatusTooManyRequests,
						Description: "service is busy",
					})
				default:
					return c.JSON(http.StatusInternalServerError, bp.Error{
						ErrorCode:   http.StatusInternalServerError,
						Description: "internal server error",
					})
				}
			}

			return nil
		}
	}
}
