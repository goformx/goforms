package contact

import "context"

// Service defines the interface for contact form business logic
type Service interface {
	Submit(ctx context.Context, sub *Submission) error
	ListSubmissions(ctx context.Context) ([]Submission, error)
	GetSubmission(ctx context.Context, id int64) (*Submission, error)
	UpdateSubmissionStatus(ctx context.Context, id int64, status Status) error
}
