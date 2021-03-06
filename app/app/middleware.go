package app

import (
	"context"
	"net/http"

	"github.com/alufhigi/http-server/db"
	"github.com/alufhigi/http-server/utils"
)

type Middleware func(http.HandlerFunc) http.HandlerFunc

func Chain(f http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {
	for _, m := range middlewares {
		f = m(f)
	}
	return f
}

func (s *Server) LoginOnly() Middleware {
	return func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			u, ok := s.IsLogin(r)
			if !ok {
				s.NotFound(w, r, http.StatusForbidden)
				return
			}
			req := r.WithContext(context.WithValue(r.Context(), "user", u))
			*r = *req
			w.Header().Set("Content-Type", "application/json")
			f(w, r)

		}
	}
}

func (s *Server) AdminOnly() Middleware {
	return func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			switch r.Context().Value("user").(type) {
			case *db.User:
				u := r.Context().Value("user").(*db.User)
				if !u.IsAdmin {
					s.NotFound(w, r, http.StatusForbidden)
					return
				}
			}
			f(w, r)

		}
	}
}

func (s *Server) Method(m string) Middleware {
	return func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.Method != m {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			f(w, r)

		}
	}
}

func (s *Server) IsLogin(r *http.Request) (*db.User, bool) {
	token, ok := r.Header["Authorization"]
	if !ok || len(token[0]) < 1 {
		return nil, false
	}
	token[0] = token[0][7:]
	userID, err := utils.ParseToken(token[0])
	if err != nil {
		return nil, false
	}
	id := userID
	u, err := s.Db.FindOneUserByID(id)
	if err != nil {
		return nil, false
	}
	return u, true

}

func (s *Server) NotFound(w http.ResponseWriter, r *http.Request, status int) {
	w.WriteHeader(status)
	if status == http.StatusNotFound {
		w.Write([]byte("404 - Page not found"))

	}
	if status == http.StatusForbidden {
		w.Write([]byte("403 - Forbidden"))
	}

}
