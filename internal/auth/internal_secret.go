package auth

import (
	"crypto/subtle"
	nethttp "net/http"

	pingoohttp "pingoo_calls/internal/http"
)

const InternalSecretHeader = "X-Pingoo-Internal-Secret"

func RequireInternalSecret(secret string, next nethttp.Handler) nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		providedSecret := r.Header.Get(InternalSecretHeader)

		if providedSecret == "" {
			pingoohttp.WriteError(w, nethttp.StatusUnauthorized, "missing internal secret")
			return
		}

		if subtle.ConstantTimeCompare([]byte(providedSecret), []byte(secret)) != 1 {
			pingoohttp.WriteError(w, nethttp.StatusUnauthorized, "invalid internal secret")
			return
		}

		next.ServeHTTP(w, r)
	})
}
