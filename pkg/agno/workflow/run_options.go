package workflow

import "github.com/rexleimo/agno-go/pkg/agno/media"

// RunOption configures workflow run behaviour.
type RunOption func(*runOptions)

type runOptions struct {
	userID         string
	sessionState   map[string]interface{}
	resumeFromStep string
	mediaPayload   []media.Attachment
	metadata       map[string]interface{}
	mediaError     error
}

// WithUserID sets the user ID for the workflow execution context.
func WithUserID(userID string) RunOption {
	return func(o *runOptions) {
		o.userID = userID
	}
}

// WithSessionState injects a session state snapshot for the execution context.
func WithSessionState(state map[string]interface{}) RunOption {
	return func(o *runOptions) {
		o.sessionState = state
	}
}

// WithResumeFrom instructs the workflow to resume from the specified step ID.
func WithResumeFrom(stepID string) RunOption {
	return func(o *runOptions) {
		o.resumeFromStep = stepID
	}
}

// WithMediaPayload attaches media payload to the execution context.
func WithMediaPayload(payload interface{}) RunOption {
	return func(o *runOptions) {
		attachments, err := media.Normalize(payload)
		if err != nil {
			o.mediaError = err
			return
		}
		o.mediaPayload = attachments
	}
}

// WithMetadata injects arbitrary metadata into the execution context.
func WithMetadata(metadata map[string]interface{}) RunOption {
	return func(o *runOptions) {
		if len(metadata) == 0 {
			return
		}
		if o.metadata == nil {
			o.metadata = make(map[string]interface{}, len(metadata))
		}
		for k, v := range metadata {
			o.metadata[k] = v
		}
	}
}

func evaluateOptions(opts []RunOption) *runOptions {
	options := &runOptions{}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(options)
	}
	if options.metadata == nil {
		options.metadata = make(map[string]interface{})
	}
	if options.mediaPayload != nil {
		options.metadata["media"] = options.mediaPayload
	}
	return options
}
