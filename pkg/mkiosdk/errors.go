package mkiosdk

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
)

// NewResponseError creates a new *ResponseError from the provided HTTP response.
// Call this when a service request returns a non-successful status code.
func NewResponseError(resp *http.Response) error {
	respErr := &ResponseError{
		StatusCode:  resp.StatusCode,
		RawResponse: resp,
	}

	// prefer the error code in the response header
	// if ec := resp.Header.Get(shared.HeaderXMSErrorCode); ec != "" {
	// 	respErr.ErrorCode = ec
	// 	return respErr
	// }

	// if we didn't get x-ms-error-code, check in the response body
	body, err := Payload(resp, nil)
	if err != nil {
		return err
	}

	if len(body) > 0 {
		if code := extractErrorCodeJSON(body); code != "" {
			respErr.ErrorCode = code
		} else if code := extractErrorCodeXML(body); code != "" {
			respErr.ErrorCode = code
		}
	}

	return respErr
}

func extractErrorCodeJSON(body []byte) string {
	var rawObj map[string]interface{}
	if err := json.Unmarshal(body, &rawObj); err != nil {
		// not a JSON object
		return ""
	}

	// check if this is a wrapped error, i.e. { "error": { ... } }
	// if so then unwrap it
	if wrapped, ok := rawObj["error"]; ok {
		unwrapped, ok := wrapped.(map[string]interface{})
		if !ok {
			return ""
		}
		rawObj = unwrapped
	} else if wrapped, ok := rawObj["odata.error"]; ok {
		// check if this a wrapped odata error, i.e. { "odata.error": { ... } }
		unwrapped, ok := wrapped.(map[string]any)
		if !ok {
			return ""
		}
		rawObj = unwrapped
	}

	// now check for the error code
	code, ok := rawObj["code"]
	if !ok {
		return ""
	}
	codeStr, ok := code.(string)
	if !ok {
		return ""
	}
	return codeStr
}

func extractErrorCodeXML(body []byte) string {
	// regular expression is much easier than dealing with the XML parser
	rx := regexp.MustCompile(`<(?:\w+:)?[c|C]ode>\s*(\w+)\s*<\/(?:\w+:)?[c|C]ode>`)
	res := rx.FindStringSubmatch(string(body))
	if len(res) != 2 {
		return ""
	}
	// first submatch is the entire thing, second one is the captured error code
	return res[1]
}

// ResponseError is returned when a request is made to a service and
// the service returns a non-success HTTP status code.
// Use errors.As() to access this type in the error chain.
type ResponseError struct {
	// ErrorCode is the error code returned by the resource provider if available.
	ErrorCode string

	// StatusCode is the HTTP status code as defined in https://pkg.go.dev/net/http#pkg-constants.
	StatusCode int

	// RawResponse is the underlying HTTP response.
	RawResponse *http.Response
}

// Error implements the error interface for type ResponseError.
// Note that the message contents are not contractual and can change over time.
func (e *ResponseError) Error() string {

	// write the request method and URL with response status code
	msg := &bytes.Buffer{}
	fmt.Fprintf(msg, "%s %s://%s%s;", e.RawResponse.Request.Method, e.RawResponse.Request.URL.Scheme, e.RawResponse.Request.URL.Host, e.RawResponse.Request.URL.Path)
	fmt.Fprintf(msg, "RESPONSE %d: %s;", e.RawResponse.StatusCode, e.RawResponse.Status)
	if e.ErrorCode != "" {
		fmt.Fprintf(msg, "ERROR CODE: %s", e.ErrorCode)
	} else {
		fmt.Fprintf(msg, "ERROR CODE UNAVAILABLE")
	}
	fmt.Fprintf(msg, ";")
	body, err := Payload(e.RawResponse, nil)
	if err != nil {
		// this really shouldn't fail at this point as the response
		// body is already cached (it was read in NewResponseError)
		fmt.Fprintf(msg, "Error reading response body: %v", err)
	} else if len(body) > 0 {
		if err := json.Indent(msg, body, "", "  "); err != nil {
			// failed to pretty-print so just dump it verbatim
			fmt.Fprint(msg, string(body))
		}
		// the standard library doesn't have a pretty-printer for XML
		fmt.Fprintln(msg)
	} else {
		fmt.Fprintf(msg, "Response contained no body")
	}
	fmt.Fprintf(msg, ";")

	return msg.String()
}

// HasStatusCode returns true if the Response's status code is one of the specified values.
// Exported as runtime.HasStatusCode().
func HasStatusCode(resp *http.Response, statusCodes ...int) bool {
	if resp == nil {
		return false
	}
	for _, sc := range statusCodes {
		if resp.StatusCode == sc {
			return true
		}
	}
	return false
}

// PayloadOptions contains the optional values for the Payload func.
// NOT exported but used by azcore.
type PayloadOptions struct {
	// BytesModifier receives the downloaded byte slice and returns an updated byte slice.
	// Use this to modify the downloaded bytes in a payload (e.g. removing a BOM).
	BytesModifier func([]byte) []byte
}

// Payload reads and returns the response body or an error.
// On a successful read, the response body is cached.
// Subsequent reads will access the cached value.
// Exported as runtime.Payload() WITHOUT the opts parameter.
func Payload(resp *http.Response, opts *PayloadOptions) ([]byte, error) {
	modifyBytes := func(b []byte) []byte { return b }
	if opts != nil && opts.BytesModifier != nil {
		modifyBytes = opts.BytesModifier
	}

	// r.Body won't be a nopClosingBytesReader if downloading was skipped
	if buf, ok := resp.Body.(*nopClosingBytesReader); ok {
		bytesBody := modifyBytes(buf.Bytes())
		buf.Set(bytesBody)
		return bytesBody, nil
	}

	bytesBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}

	bytesBody = modifyBytes(bytesBody)
	resp.Body = &nopClosingBytesReader{s: bytesBody}
	return bytesBody, nil
}

// PayloadDownloaded returns true if the response body has already been downloaded.
// This implies that the Payload() func above has been previously called.
// NOT exported but used by azcore.
func PayloadDownloaded(resp *http.Response) bool {
	_, ok := resp.Body.(*nopClosingBytesReader)
	return ok
}

// nopClosingBytesReader is an io.ReadSeekCloser around a byte slice.
// It also provides direct access to the byte slice to avoid rereading.
type nopClosingBytesReader struct {
	s []byte
	i int64
}

// Bytes returns the underlying byte slice.
func (r *nopClosingBytesReader) Bytes() []byte {
	return r.s
}

// Close implements the io.Closer interface.
func (*nopClosingBytesReader) Close() error {
	return nil
}

// Read implements the io.Reader interface.
func (r *nopClosingBytesReader) Read(b []byte) (n int, err error) {
	if r.i >= int64(len(r.s)) {
		return 0, io.EOF
	}
	n = copy(b, r.s[r.i:])
	r.i += int64(n)
	return
}

// Set replaces the existing byte slice with the specified byte slice and resets the reader.
func (r *nopClosingBytesReader) Set(b []byte) {
	r.s = b
	r.i = 0
}

// Seek implements the io.Seeker interface.
func (r *nopClosingBytesReader) Seek(offset int64, whence int) (int64, error) {
	var i int64
	switch whence {
	case io.SeekStart:
		i = offset
	case io.SeekCurrent:
		i = r.i + offset
	case io.SeekEnd:
		i = int64(len(r.s)) + offset
	default:
		return 0, errors.New("nopClosingBytesReader: invalid whence")
	}
	if i < 0 {
		return 0, errors.New("nopClosingBytesReader: negative position")
	}
	r.i = i
	return i, nil
}
