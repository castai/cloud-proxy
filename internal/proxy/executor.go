package proxy

//
//import (
//	"fmt"
//	"net/http"
//
//	"github.com/castai/cloud-proxy/internal/gcpauth"
//)
//
//type Executor struct {
//	credentialsSrc gcpauth.GCPCredentialsSource
//	client         *http.Client
//}
//
//func (e *Executor) DoRequest(request *http.Request) (*http.Response, error) {
//
//	credentials, err := e.credentialsSrc.GetDefaultCredentials()
//	if err != nil {
//		return nil, fmt.Errorf("cannot load GCP credentials: %w", err)
//	}
//
//	token, err := credentials.TokenSource.Token()
//	if err != nil {
//		return nil, fmt.Errorf("cannot get access token from src (%T): %w", credentials.TokenSource, err)
//	}
//
//	req, err := http.NewRequest(request.Method, request.URL.String(), request.Body)
//	return nil, nil
//}
//
////creds := getGCPCredential()
////
////token, err := creds.TokenSource.Token()
////if err != nil {
////return nil, fmt.Errorf("failed to get auth token: %w", err)
////}
////
////req, err := http.NewRequest(protoReq.Method, protoReq.Url, bytes.NewReader(protoReq.Body))
////if err != nil {
////return nil, fmt.Errorf("failed to create proxy http request: %w", err)
////}
////
////// Set the authorize header manually
////req.Header.Add("Authorization", "Bearer "+token.AccessToken)
////for header, val := range protoReq.Headers {
////if strings.ToLower(header) == "authorization" {
////continue
////}
////req.Header.Add(header, val)
////}
////
////resp, err := e.client.Do(req)
////if err != nil {
////return nil, fmt.Errorf("unexpected err for %+v: %w", protoReq, err)
////}
