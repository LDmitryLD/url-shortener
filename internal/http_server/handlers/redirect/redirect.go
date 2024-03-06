package redirect

import (
	"errors"
	"net/http"

	"github.com/LDmitryLD/url-shortener/internal/lib/api/response"
	"github.com/LDmitryLD/url-shortener/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"go.uber.org/zap"
)

//go:generate go run github.com/vektra/mockery/v2@v2.35.4 --name=URLGetter
type URLGetter interface {
	GetURL(alias string) (string, error)
}

func New(log *zap.Logger, urlGetter URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.redirect.New"

		log = log.With(
			zap.String("op", op),
			zap.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("alias is empty")

			render.JSON(w, r, response.Error("invalid request"))

			return
		}

		resURL, err := urlGetter.GetURL(alias)
		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("url not found", zap.String("alias", alias))

			render.JSON(w, r, response.Error("not found"))

			return
		}
		if err != nil {
			log.Error("faild to get url", zap.Error(err))

			render.JSON(w, r, response.Error("internal error"))

			return
		}

		log.Info("got url", zap.String("url", resURL))

		http.Redirect(w, r, resURL, http.StatusFound)
	}
}
