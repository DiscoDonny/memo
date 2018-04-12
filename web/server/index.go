package server

import (
	"git.jasonc.me/main/memo/app/auth"
	"git.jasonc.me/main/memo/app/db"
	"git.jasonc.me/main/memo/app/profile"
	"git.jasonc.me/main/memo/app/res"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/web"
	"net/http"
)

var (
	indexRoute = web.Route{
		Pattern: res.UrlIndex,
		Handler: func(r *web.Response) {
			if ! auth.IsLoggedIn(r.Session.CookieId) {
				r.Render()
				return
			}
			user, err := auth.GetSessionUser(r.Session.CookieId)
			if err != nil {
				r.Error(jerr.Get("error getting session user", err), http.StatusInternalServerError)
				return
			}
			key, err := db.GetKeyForUser(user.Id)
			if err != nil {
				r.Error(jerr.Get("error getting key for user", err), http.StatusInternalServerError)
				return
			}
			r.Helper["Key"] = key

			pf, err := profile.GetProfileAndSetBalances(key.PkHash)
			if err != nil {
				r.Error(jerr.Get("error getting profile for hash", err), http.StatusInternalServerError)
				return
			}
			pf.Self = true
			r.Helper["Profile"] = pf

			posts, err := db.GetPostsForPkHash(key.PkHash)
			if err != nil {
				r.Error(jerr.Get("error getting posts for hash", err), http.StatusInternalServerError)
				return
			}
			r.Helper["Posts"] = posts
			r.RenderTemplate("dashboard")
		},
	}
)
