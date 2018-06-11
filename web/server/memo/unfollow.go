package memo

import (
	"bytes"
	"fmt"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/web"
	"github.com/memocash/memo/app/auth"
	"github.com/memocash/memo/app/bitcoin/transaction"
	"github.com/memocash/memo/app/bitcoin/transaction/build"
	"github.com/memocash/memo/app/bitcoin/wallet"
	"github.com/memocash/memo/app/db"
	"github.com/memocash/memo/app/mutex"
	"github.com/memocash/memo/app/profile"
	"github.com/memocash/memo/app/res"
	"net/http"
)

var unfollowRoute = web.Route{
	Pattern:    res.UrlMemoUnfollow + "/" + urlAddress.UrlPart(),
	NeedsLogin: true,
	Handler: func(r *web.Response) {
		addressString := r.Request.GetUrlNamedQueryVariable(urlAddress.Id)
		address := wallet.GetAddressFromString(addressString)
		pkHash := address.GetScriptAddress()
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
		if bytes.Equal(key.PkHash, pkHash) {
			r.SetRedirect(res.GetUrlWithBaseUrl(res.UrlIndex, r))
			return
		}
		hasSpendableTxOut, err := db.HasSpendable(key.PkHash)
		if err != nil {
			r.Error(jerr.Get("error getting spendable tx out", err), http.StatusInternalServerError)
			return
		}
		if ! hasSpendableTxOut {
			r.SetRedirect(res.UrlNeedFunds)
			return
		}

		pf, err := profile.GetProfile(pkHash, key.PkHash)
		if err != nil {
			r.Error(jerr.Get("error getting profile for hash", err), http.StatusInternalServerError)
			return
		}

		canFollow, err := profile.CanFollow(pkHash, key.PkHash)
		if err != nil {
			r.Error(jerr.Get("error getting can follow", err), http.StatusInternalServerError)
			return
		}
		if canFollow {
			r.Error(jerr.New("unable to unfollow user"), http.StatusUnprocessableEntity)
			return
		}
		r.Helper["Profile"] = pf
		r.RenderTemplate(res.UrlMemoUnfollow)
	},
}

var unfollowSubmitRoute = web.Route{
	Pattern:     res.UrlMemoUnfollowSubmit,
	NeedsLogin:  true,
	CsrfProtect: true,
	Handler: func(r *web.Response) {
		addressString := r.Request.GetFormValue("address")
		followAddress := wallet.GetAddressFromString(addressString)
		if followAddress.GetEncoded() != addressString {
			r.Error(jerr.New("error parsing address"), http.StatusUnprocessableEntity)
			return
		}
		password := r.Request.GetFormValue("password")
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

		privateKey, err := key.GetPrivateKey(password)
		if err != nil {
			r.Error(jerr.Get("error getting private key", err), http.StatusUnauthorized)
			return
		}

		pkHash := privateKey.GetPublicKey().GetAddress().GetScriptAddress()
		mutex.Lock(pkHash)

		tx, err := build.UnfollowUser(followAddress.GetScriptAddress(), privateKey)
		if err != nil {
			mutex.Unlock(pkHash)
			r.Error(jerr.Get("error building unfollow tx", err), http.StatusInternalServerError)
			return
		}

		fmt.Println(transaction.GetTxInfo(tx))
		transaction.QueueTx(tx)
		r.Write(tx.TxHash().String())
	},
}
