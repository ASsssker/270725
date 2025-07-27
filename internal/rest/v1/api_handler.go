package v1

import (
	bp "270725/internal/rest/v1/boileplate"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
)

func (h *Handler) GetAPI(c echo.Context) error {
	swagger, err := bp.GetSwagger()
	if err != nil {
		return fmt.Errorf("failed to get swagger error: %w", err)
	}

	jsSwagger, err := swagger.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal swagger: %w", err)
	}

	return c.JSON(http.StatusOK, jsSwagger)
}
