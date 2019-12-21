// +build go1.12

package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/metadata"
)

const (
	sessionCookieName = "sessionid"
	stateCookieName   = "oidcstate"
	HttpCookiePath    = "/"
)

// cookieMapper is mapping an outgoing session to a session cookie
// This function is used on the login call
// Information processed here are generated by our own service
/*
func cookieMapper(ctx context.Context, w http.ResponseWriter, resp proto.Message) error {
	_, ok := resp.(*Empty) // logout
	if ok {
		return nil
	}

	session, ok := resp.(*SessionData) // login we map.
	if !ok {
		return errors.New("Cannot cast to Session")
	}

	cookie := &http.Cookie{
		Name:  sessionCookieName,
		Value: session.Id,

		Path:     HttpCookiePath,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}

	// XXX Create cookie the wrong way
	//cookie := fmt.Sprintf(HttpCookieFormat, HttpCookieToken, session.Id)
	// we must use Cookie creation API from net/http and set in the
	// header
	http.SetCookie(w, cookie)
	// empty the content it's in the header.
	//session.Id = ""
	return nil
}
*/

// ON ERROR it does NOT go through here, so we cannot interfere
func cookieOrRedirectMapper(ctx context.Context, w http.ResponseWriter, resp proto.Message) error {
	fmt.Printf("EXEC COOKIE BLA\n")

	switch v := resp.(type) {
	case *Empty:
		/*
			case *RedirectData:
				w.Header().Add("Location", v.Url)
			case *RedirectSessionData:
		*/
	case *SessionBackend:
		var cookie *http.Cookie
		sessioncookie := v.GetCookieSession()
		url := v.GetUrl()
		fmt.Printf("BACKEND sessionc: %s\n", sessioncookie)

		// session id cookie
		if len(sessioncookie) > 0 {
			cookie = &http.Cookie{
				Name:  sessionCookieName,
				Value: sessioncookie,

				Path:     HttpCookiePath,
				Secure:   true,
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
			}
			w.Header().Add("Location", url)
		}
		http.SetCookie(w, cookie)
		return ErrRedirect
	case *SessionIdp:
		var cookie *http.Cookie
		statecookie := v.GetCookieState()
		cookiepath := v.GetCookiePath()
		url := v.GetUrl()
		fmt.Printf("OIDC SESSION statec: %s pathc: %s  url: %s\n",
			statecookie,
			cookiepath,
			url)
		// session id cookie
		if len(statecookie) > 0 && len(url) > 0 && len(cookiepath) > 0 { // this is the first redirect
			cookie = &http.Cookie{
				Name:  stateCookieName,
				Value: statecookie,

				//Path:     HttpCookiePath,
				Path:     cookiepath,
				Secure:   true,
				HttpOnly: true,
				// XXX super odd, but with chrome on strict mode
				// the first request does NOT come with the
				// cookie previously set.
				//SameSite: http.SameSiteStrictMode,
				SameSite: http.SameSiteLaxMode,
				// XXX TODO: if you do that it becomes a permanent cookie and we WANT to remain just a session
				//MaxAge:   int((5 * time.Minute).Seconds()),
				//Expires:  time.Now().Add(5 * time.Minute),

				// TODO
				// need Expires
				// need Domain
			}
			w.Header().Add("Cache-Control", "no-cache, no-store, max-age=0, must-revalidate")
			w.Header().Add("Location", url)
		}
		http.SetCookie(w, cookie)
		return ErrRedirect
	}

	// XXX Create cookie the wrong way
	//cookie := fmt.Sprintf(HttpCookieFormat, HttpCookieToken, session.Id)
	// we must use Cookie creation API from net/http and set in the
	// header
	// empty the content it's in the header.
	//session.Id = ""
	return nil
}

func headerRemover(headerName string) (mdName string, ok bool) {
	//header := strings.ToLower(headerName)
	//headerLen := len(headerName)

	// This fonction just prevent those headers out
	// XXX we need to check if any header received on the gateway can tamper
	// with grpc behavior
	/*
	   if headerLen == 26 && strings.Compare(header, "grpc-metadata-content-type") == 0 {
	           return "", false
	   }
	*/

	/*
		if headerLen == len(rpc.GrpcMdContentType) && strings.EqualFold(headerName, rpc.GrpcMdContentType) {
			return "", false
		}
	*/

	return headerName, true
}

func headerToMetadata(ctx context.Context, r *http.Request) metadata.MD {
	mdmap := make(map[string]string)

	// XXX needs input validation happening here..
	// these data should be SHORT.

	cookieSession, err := r.Cookie(sessionCookieName)
	if err == nil {
		mdmap[sessionCookieName] = cookieSession.Value
	}
	cookieState, err := r.Cookie(stateCookieName)
	// 8k is more than enough..
	if err == nil && len(cookieState.Value) < 8192 {
		mdmap[stateCookieName] = cookieState.Value
	}

	return metadata.New(mdmap)
}
