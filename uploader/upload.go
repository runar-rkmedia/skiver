package uploader

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strings"
	"text/template"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/runar-rkmedia/go-common/logger"
	"github.com/runar-rkmedia/skiver/types"
)

type s3Uploader struct {
	ID          string
	EndpointURL url.URL
	S3UploaderOptions
	ForcePathStyle bool
	Provider       *Provider
	L              logger.AppLogger
	privateKey     string
}
type S3UploaderOptions struct {
	Endpoint,
	Region,
	Bucket,
	ProviderName,
	AccessKey string
	ForcePathStyle bool
	UrlFormat      string
	CacheControl   string
}

func NewS3Uplaoder(
	l logger.AppLogger,
	Identifier string,
	options S3UploaderOptions,
	privateKey string,
	// Endpoint for the s3-compatible service
) *s3Uploader {
	L := logger.With(l.With().
		Str("endpoint", options.Endpoint).
		Str("bucket", options.Bucket).
		Str("identifier", Identifier).
		Logger())

	if options.Bucket == "" {
		l.Fatal().Msg("Bucket is required")
	}
	if options.Endpoint == "" {
		l.Fatal().Msg("Endpoint is required")
	}
	if options.ProviderName == "" {
		l.Fatal().Msg("ProviderName is required")
	}
	if options.Region == "" {
		l.Fatal().Msg("Region is required")
	}
	if options.AccessKey == "" {
		l.Fatal().Msg("AccessKey is required")
	}
	if privateKey == "" {
		l.Fatal().Msg("privateKey is required")
	}
	endpointUrl, err := url.Parse(options.Endpoint)
	if err != nil {
		l.Fatal().Err(err).Msg("endpoint is not valid")
	}

	s := &s3Uploader{
		L:                 L,
		EndpointURL:       *endpointUrl,
		S3UploaderOptions: options,
		ID:                Identifier,
		privateKey:        privateKey,
	}

	if options.UrlFormat != "" {
		s.Provider = &Provider{
			Name:      options.ProviderName,
			UrlFormat: options.ProviderName,
		}
	}
	if s.Provider == nil {
		s.Provider = findProvider(options.Endpoint)
	}
	return s
}

type FileUploader interface {
	AddPublicFile(key string, r io.ReadSeeker, size int64, contentType string, contentDisposition string) (types.UploadMeta, error)
	AddPublicFileWithAliases(keys []string, r io.ReadSeeker, size int64, contentType string, contentDisposition string) ([]types.UploadMeta, error)
	UrlForFile(objectID string) (string, error)
	Identifier() string
}

func (su *s3Uploader) Identifier() string {
	return su.ID
}

// TODO: Vi må lagre i databasen om vi har lastet opp til instansen, for å hindre at vi laster opp igjen og igjen.
// Planen er at hver gang det lages en snapshot, blir snapshotten lastet opp, og hvis de er semvar, blir det lastet opp opp til fire kopier:
// Eksempel for snapshot "v3.2.1-beta": ["v3", "v3.2", "v3.2.1", "v3.2.1-beta"].
// Det er ikke nødvendig med latest-opplasting, siden dette vil være mest for utviklere.

func (su *s3Uploader) newSession() (*session.Session, error) {

	if su.L.HasDebug() {
		su.L.Debug().Msg("Getting session")
	}
	s, err := session.NewSession(&aws.Config{
		Endpoint:         aws.String(su.Endpoint),
		S3ForcePathStyle: aws.Bool(su.ForcePathStyle),
		Credentials: credentials.NewCredentials(&credentials.StaticProvider{
			Value: credentials.Value{
				AccessKeyID:     su.AccessKey,
				SecretAccessKey: su.privateKey,
				SessionToken:    "",
				ProviderName:    "local-config",
			},
		}),
		CredentialsChainVerboseErrors: aws.Bool(true),
		Region:                        aws.String(su.Region)})

	if err != nil {
		su.L.Error().Err(err).Msg("Failed to create session for s3")
		return s, err
	}

	if su.L.HasDebug() {
		su.L.Debug().Msg("Got a session, starting upload")
	}

	return s, err
}

func (su *s3Uploader) getClient() (*s3.S3, error) {

	s, err := su.newSession()
	if err != nil {
		return nil, err
	}
	return s3.New(s), nil

}
func (su *s3Uploader) GetFileInfo(key string) error {
	client, err := su.getClient()
	if err != nil {
		return err
	}

	input := &s3.GetObjectAttributesInput{
		Bucket: aws.String(su.Bucket),
		Key:    aws.String(su.Bucket),
	}
	client.GetObjectAttributes(input)

	return nil
}

type Provider struct {
	Name          string
	endpointRegex *regexp.Regexp
	UrlFormat     string
}

func (p Provider) String() string { return p.Name }

var (
	// TODO: Ensure this url is correct. This is just a placeholder for now.
	ProviderBackBlaze = Provider{"BackBlazeB@", regexp.MustCompile(`.*backblazeb2\.com\/?$`), "{{.EndpointURL.Scheme}}://{{.Bucket}}.{{.EndpointURL.Host}}/{{.Object}}"}
	providers         = []Provider{ProviderBackBlaze}
)

func findProvider(endpoint string) *Provider {
	for _, p := range providers {
		if p.endpointRegex.MatchString(endpoint) {
			return &p
		}
	}
	return nil
}

func (su *s3Uploader) UrlForFile(objectID string) (string, error) {

	if su.Provider != nil && su.Provider.UrlFormat != "" {

		tmpl, err := template.New("").Parse(su.Provider.UrlFormat)
		if err != nil {
			return "", err
		}
		if err != nil {
			return "", err
		}
		w := bytes.Buffer{}
		data := struct {
			EndpointURL                      url.URL
			Endpoint, Region, Bucket, Object string
		}{
			EndpointURL: su.EndpointURL,
			Endpoint:    su.Endpoint,
			Region:      su.Region,
			Bucket:      su.Bucket,
			Object:      objectID,
		}
		tmpl.Execute(&w, data)

		url := w.String()
		if url != "" {
			return url, nil
		}

	}

	// Generic url for aws
	client, err := su.getClient()
	if err != nil {
		return "", err
	}

	req, _ := client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(su.Bucket),
		Key:    aws.String(su.Bucket),
	})
	url := req.HTTPRequest.URL.String()

	return url, nil

}
func (su *s3Uploader) AddPublicFileWithAliases(keys []string, r io.ReadSeeker, size int64, contentType string, contentDisposition string) ([]types.UploadMeta, error) {
	length := len(keys)
	if length == 0 {
		return nil, fmt.Errorf("received zero keys")
	}

	first, err := su.AddPublicFile(keys[0], r, size, contentType, contentDisposition)
	if err != nil {
		return nil, err
	}
	m := make([]types.UploadMeta, length)
	m[0] = first
	if length == 1 {
		return m, nil
	}
	client, err := su.getClient()
	if err != nil {
		return m, err
	}
	for i := 1; i < length; i++ {
		input := s3.CopyObjectInput{
			Bucket:            aws.String(su.Bucket),
			CopySource:        aws.String(su.Bucket + "/" + keys[0]),
			Key:               aws.String(keys[i]),
			MetadataDirective: aws.String("COPY"),
		}
		_, err := client.CopyObject(&input)
		if err != nil {
			return m, err
		}
		if su.L.HasDebug() {
			su.L.Debug().Msg("Made a copy of the file")
		}
		m[i] = types.UploadMeta{
			ID:           keys[i],
			Parent:       su.Bucket,
			ProviderID:   su.Identifier(),
			ProviderName: su.ProviderName,
			Size:         size,
		}
		url, err := su.UrlForFile(keys[i])
		if err != nil {
			return m, err
		}
		m[i].URL = url
	}

	return m, err

}
func (su *s3Uploader) AddPublicFile(key string, r io.ReadSeeker, size int64, contentType string, contentDisposition string) (types.UploadMeta, error) {
	if key == "" {
		return types.UploadMeta{}, fmt.Errorf("key was empty")
	}
	// We do not want to upload snapshots that are empty, since that could overwrite existing content with empty content
	// A file is considered empty if it holds not useful information, like an empty json-object like '{}' or '[]' etc
	if size == 0 {
		return types.UploadMeta{}, fmt.Errorf("Empty file, refusing to upload")
	}
	if size <= 4 {
		buf := new(strings.Builder)
		_, err := io.Copy(buf, r)
		if err != nil {
			return types.UploadMeta{}, fmt.Errorf("Failed to verify file for non-null content: %w", err)
		}
		s := buf.String()
		switch s {
		case "null":
			return types.UploadMeta{}, fmt.Errorf("Empty file (null-string!), refusing to upload")
		case "{}":
			return types.UploadMeta{}, fmt.Errorf("Empty file ({}-string!), refusing to upload")
		case "[]":
			return types.UploadMeta{}, fmt.Errorf("Empty file ([]-string!), refusing to upload")

		}
		r.Seek(0, io.SeekStart)
	}
	um := types.UploadMeta{
		ID:           key,
		Parent:       su.Bucket,
		ProviderID:   su.Identifier(),
		ProviderName: su.ProviderName,
		Size:         size,
	}
	url, err := su.UrlForFile(key)
	if err != nil {
		return um, err
	}
	um.URL = url
	l := logger.With(su.L.With().
		Str("key", key).
		Int64("size", size).
		Str("contentType", contentType).
		Str("contentDisposition", contentDisposition).
		Logger())

	client, err := su.getClient()
	if err != nil {
		return um, err
	}

	putInput := s3.PutObjectInput{
		Bucket: aws.String(su.Bucket),
		Key:    aws.String(key),
		// ACL:                  aws.String("private"),
		Body:                 r,
		ContentLength:        aws.Int64(size),
		ContentType:          aws.String(contentType),
		ContentDisposition:   aws.String(contentDisposition),
		ServerSideEncryption: aws.String("AES256"),
	}
	if su.CacheControl != "" {
		putInput.CacheControl = aws.String(su.CacheControl)
	}
	_, err = client.PutObject(&putInput)

	if err != nil {
		l.Error().Err(err).Interface("input", putInput).Msg("Failed to upload to s3")
		return um, err
	}
	l.Info().Interface("uploadMeta", um).Msg("File uploaded successfully to s3")
	return um, nil

}
