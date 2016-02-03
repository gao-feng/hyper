package local

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/docker/distribution/digest"
	"github.com/docker/docker/builder/dockerfile"
	"github.com/docker/docker/pkg/ioutils"
	"github.com/docker/docker/pkg/streamformatter"
	"github.com/docker/docker/reference"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/container"
	"github.com/hyperhq/hyper/server/httputils"
	"golang.org/x/net/context"
)

// Creates an image from Pull or from Import
func (s *router) postImagesCreate(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := httputils.ParseForm(r); err != nil {
		return err
	}

	var (
		image   = r.Form.Get("imageName")
		repo    = r.Form.Get("repo")
		tag     = r.Form.Get("tag")
		message = r.Form.Get("message")
	)
	authEncoded := r.Header.Get("X-Registry-Auth")
	authConfig := &types.AuthConfig{}
	if authEncoded != "" {
		authJSON := base64.NewDecoder(base64.URLEncoding, strings.NewReader(authEncoded))
		if err := json.NewDecoder(authJSON).Decode(authConfig); err != nil {
			// for a pull it is not an error if no auth was given
			// to increase compatibility with the existing api it is defaulting to be empty
			authConfig = &types.AuthConfig{}
		}
	}

	var (
		err    error
		output = ioutils.NewWriteFlusher(w)
	)
	defer output.Close()

	w.Header().Set("Content-Type", "application/json")

	if image != "" { //pull
		// Special case: "pull -a" may send an image name with a
		// trailing :. This is ugly, but let's not break API
		// compatibility.
		image = strings.TrimSuffix(image, ":")

		var ref reference.Named
		ref, err = reference.ParseNamed(image)
		if err == nil {
			if tag != "" {
				// The "tag" could actually be a digest.
				var dgst digest.Digest
				dgst, err = digest.ParseDigest(tag)
				if err == nil {
					ref, err = reference.WithDigest(ref, dgst)
				} else {
					ref, err = reference.WithTag(ref, tag)
				}
			}
			if err == nil {
				metaHeaders := map[string][]string{}
				for k, v := range r.Header {
					if strings.HasPrefix(k, "X-Meta-") {
						metaHeaders[k] = v
					}
				}

				err = s.daemon.PullImage(ref, metaHeaders, authConfig, output)
			}
		}
	} else { //import
		var newRef reference.Named
		if repo != "" {
			var err error
			newRef, err = reference.ParseNamed(repo)
			if err != nil {
				return err
			}

			if _, isCanonical := newRef.(reference.Canonical); isCanonical {
				return errors.New("cannot import digest reference")
			}

			if tag != "" {
				newRef, err = reference.WithTag(newRef, tag)
				if err != nil {
					return err
				}
			}
		}

		src := r.Form.Get("fromSrc")

		// 'err' MUST NOT be defined within this block, we need any error
		// generated from the download to be available to the output
		// stream processing below
		var newConfig *container.Config
		newConfig, err = dockerfile.BuildFromConfig(&container.Config{}, r.Form["changes"])
		if err != nil {
			return err
		}

		err = s.daemon.ImportImage(src, newRef, message, r.Body, output, newConfig)
	}
	if err != nil {
		if !output.Flushed() {
			return err
		}
		sf := streamformatter.NewJSONStreamFormatter()
		output.Write(sf.FormatError(err))
	}

	return nil
}

func (s *router) postImagesPush(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	metaHeaders := map[string][]string{}
	for k, v := range r.Header {
		if strings.HasPrefix(k, "X-Meta-") {
			metaHeaders[k] = v
		}
	}
	if err := httputils.ParseForm(r); err != nil {
		return err
	}
	authConfig := &types.AuthConfig{}

	authEncoded := r.Header.Get("X-Registry-Auth")
	if authEncoded != "" {
		// the new format is to handle the authConfig as a header
		authJSON := base64.NewDecoder(base64.URLEncoding, strings.NewReader(authEncoded))
		if err := json.NewDecoder(authJSON).Decode(authConfig); err != nil {
			// to increase compatibility to existing api it is defaulting to be empty
			authConfig = &types.AuthConfig{}
		}
	} else {
		// the old format is supported for compatibility if there was no authConfig header
		if err := json.NewDecoder(r.Body).Decode(authConfig); err != nil {
			return fmt.Errorf("Bad parameters and missing X-Registry-Auth: %v", err)
		}
	}

	ref, err := reference.ParseNamed(vars["name"])
	if err != nil {
		return err
	}
	tag := r.Form.Get("tag")
	if tag != "" {
		// Push by digest is not supported, so only tags are supported.
		ref, err = reference.WithTag(ref, tag)
		if err != nil {
			return err
		}
	}

	output := ioutils.NewWriteFlusher(w)
	defer output.Close()

	w.Header().Set("Content-Type", "application/json")

	if err := s.daemon.PushImage(ref, metaHeaders, authConfig, output); err != nil {
		if !output.Flushed() {
			return err
		}
		sf := streamformatter.NewJSONStreamFormatter()
		output.Write(sf.FormatError(err))
	}
	return nil
}

func (s *router) deleteImages(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := httputils.ParseForm(r); err != nil {
		return err
	}

	name := r.Form.Get("imageId")

	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("image name cannot be blank")
	}

	force := httputils.BoolValue(r, "force")
	prune := !httputils.BoolValue(r, "noprune")

	env, err := s.daemon.CmdImageDelete(name, force, prune)
	if err != nil {
		return err
	}

	return env.WriteJSON(w, http.StatusOK)
}

func (s *router) getImagesJSON(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := httputils.ParseForm(r); err != nil {
		return err
	}

	// FIXME: The filter parameter could just be a match filter
	env, err := s.daemon.CmdImages(r.Form.Get("filters"), r.Form.Get("filter"), httputils.BoolValue(r, "all"))
	if err != nil {
		return err
	}

	return env.WriteJSON(w, http.StatusOK)
}
