package application

import (
	"context"
	"errors"
	"testing"

	"github.com/samber/mo"
	"go.mongodb.org/mongo-driver/v2/bson"

	"todoe/domain/task/domain"
	"todoe/domain/task/port"
)

// ── fakes ────────────────────────────────────────────────────────────

type fakeRepo struct {
	findByIDResult mo.Result[domain.Task]
	findByIDCalls  []bson.ObjectID

	findAllResult mo.Result[[]domain.Task]
	findAllCalls  int

	saveResult mo.Result[struct{}]
	saveCalls  []domain.Task

	updateResult mo.Result[struct{}]
	updateCalls  []struct {
		id     bson.ObjectID
		status domain.Status
	}
}

func (r *fakeRepo) Save(ctx context.Context, task domain.Task) mo.Result[struct{}] {
	r.saveCalls = append(r.saveCalls, task)
	return r.saveResult
}
func (r *fakeRepo) FindAll(ctx context.Context) mo.Result[[]domain.Task] {
	r.findAllCalls++
	return r.findAllResult
}
func (r *fakeRepo) FindByID(ctx context.Context, id bson.ObjectID) mo.Result[domain.Task] {
	r.findByIDCalls = append(r.findByIDCalls, id)
	return r.findByIDResult
}
func (r *fakeRepo) UpdateStatus(ctx context.Context, id bson.ObjectID, status domain.Status) mo.Result[struct{}] {
	r.updateCalls = append(r.updateCalls, struct {
		id     bson.ObjectID
		status domain.Status
	}{id, status})
	return r.updateResult
}

type publishedEvent struct {
	name    string
	payload any
}

type fakePublisher struct {
	events []publishedEvent
}

func (p *fakePublisher) Publish(name string, payload any) {
	p.events = append(p.events, publishedEvent{name, payload})
}

var (
	_ port.Repository = (*fakeRepo)(nil)
	_ port.Publisher  = (*fakePublisher)(nil)
)

// ── tests ────────────────────────────────────────────────────────────

func TestService_CreateTask_RejectsEmptyTitle(t *testing.T) {
	repo := &fakeRepo{}
	pub := &fakePublisher{}
	svc := NewService(repo, pub)

	res := svc.CreateTask(context.Background(), "")

	if !res.IsError() {
		t.Fatalf("expected error, got value")
	}
	if !errors.Is(res.Error(), ErrInvalidTitle) {
		t.Fatalf("expected ErrInvalidTitle, got %v", res.Error())
	}
	if len(pub.events) != 0 {
		t.Fatalf("expected no events, got %d", len(pub.events))
	}
	if len(repo.saveCalls) != 0 {
		t.Fatalf("expected no Save calls, got %d", len(repo.saveCalls))
	}
}

func TestService_CreateTask_PublishesCreatedEventWithPendingStatus(t *testing.T) {
	repo := &fakeRepo{}
	pub := &fakePublisher{}
	svc := NewService(repo, pub)

	res := svc.CreateTask(context.Background(), "ship the thing")

	if res.IsError() {
		t.Fatalf("unexpected error: %v", res.Error())
	}
	task := res.MustGet()
	if task.Title != "ship the thing" {
		t.Errorf("title=%q, want %q", task.Title, "ship the thing")
	}
	if task.Status != domain.StatusPending {
		t.Errorf("status=%v, want %v", task.Status, domain.StatusPending)
	}
	if task.ID.IsZero() {
		t.Errorf("expected non-zero ID")
	}
	if task.CreatedAt.IsZero() {
		t.Errorf("expected non-zero CreatedAt")
	}

	if len(pub.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(pub.events))
	}
	ev := pub.events[0]
	if ev.name != domain.EventCreated {
		t.Errorf("event name=%q, want %q", ev.name, domain.EventCreated)
	}
	payload, ok := ev.payload.(domain.CreatedPayload)
	if !ok {
		t.Fatalf("payload type=%T, want domain.CreatedPayload", ev.payload)
	}
	if payload.Task != task {
		t.Errorf("payload Task != returned task\n got=%+v\nwant=%+v", payload.Task, task)
	}
}

func TestService_GetTask_DelegatesToRepoFindByID(t *testing.T) {
	id := bson.NewObjectID()
	expected := domain.Task{ID: id, Title: "x", Status: domain.StatusPending}
	repo := &fakeRepo{findByIDResult: mo.Ok(expected)}
	svc := NewService(repo, &fakePublisher{})

	res := svc.GetTask(context.Background(), id)

	if res.IsError() {
		t.Fatalf("unexpected error: %v", res.Error())
	}
	if res.MustGet() != expected {
		t.Errorf("got %+v, want %+v", res.MustGet(), expected)
	}
	if len(repo.findByIDCalls) != 1 || repo.findByIDCalls[0] != id {
		t.Errorf("expected one FindByID(%v), got %+v", id, repo.findByIDCalls)
	}
}

func TestService_ListTasks_DelegatesToRepoFindAll(t *testing.T) {
	want := []domain.Task{{Title: "a"}, {Title: "b"}}
	repo := &fakeRepo{findAllResult: mo.Ok(want)}
	svc := NewService(repo, &fakePublisher{})

	res := svc.ListTasks(context.Background())

	if res.IsError() {
		t.Fatalf("unexpected error: %v", res.Error())
	}
	if len(res.MustGet()) != len(want) {
		t.Errorf("len=%d, want %d", len(res.MustGet()), len(want))
	}
	if repo.findAllCalls != 1 {
		t.Errorf("expected 1 FindAll call, got %d", repo.findAllCalls)
	}
}

func TestService_ChangeStatus_RejectsInvalidStatus(t *testing.T) {
	repo := &fakeRepo{}
	pub := &fakePublisher{}
	svc := NewService(repo, pub)

	res := svc.ChangeStatus(context.Background(), bson.NewObjectID(), domain.Status("weird"))

	if !res.IsError() {
		t.Fatalf("expected error, got value")
	}
	if !errors.Is(res.Error(), ErrInvalidStatus) {
		t.Fatalf("expected ErrInvalidStatus, got %v", res.Error())
	}
	if len(repo.findByIDCalls) != 0 {
		t.Errorf("expected no FindByID call, got %d", len(repo.findByIDCalls))
	}
	if len(pub.events) != 0 {
		t.Errorf("expected no publish, got %d", len(pub.events))
	}
}

func TestService_ChangeStatus_PropagatesNotFound(t *testing.T) {
	notFound := errors.New("not found")
	repo := &fakeRepo{findByIDResult: mo.Err[domain.Task](notFound)}
	pub := &fakePublisher{}
	svc := NewService(repo, pub)

	res := svc.ChangeStatus(context.Background(), bson.NewObjectID(), domain.StatusInProgress)

	if !res.IsError() {
		t.Fatalf("expected error, got value")
	}
	if !errors.Is(res.Error(), notFound) {
		t.Fatalf("expected wrapped not-found error, got %v", res.Error())
	}
	if len(pub.events) != 0 {
		t.Errorf("expected no publish, got %d", len(pub.events))
	}
}

func TestService_ChangeStatus_PublishesEventAndReturnsUpdatedTask(t *testing.T) {
	id := bson.NewObjectID()
	initial := domain.Task{ID: id, Title: "x", Status: domain.StatusPending}
	repo := &fakeRepo{findByIDResult: mo.Ok(initial)}
	pub := &fakePublisher{}
	svc := NewService(repo, pub)

	res := svc.ChangeStatus(context.Background(), id, domain.StatusInProgress)

	if res.IsError() {
		t.Fatalf("unexpected error: %v", res.Error())
	}
	got := res.MustGet()
	if got.Status != domain.StatusInProgress {
		t.Errorf("status=%v, want %v", got.Status, domain.StatusInProgress)
	}
	if got.ID != id {
		t.Errorf("id=%v, want %v", got.ID, id)
	}

	if len(pub.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(pub.events))
	}
	ev := pub.events[0]
	if ev.name != domain.EventStatusChanged {
		t.Errorf("event name=%q, want %q", ev.name, domain.EventStatusChanged)
	}
	payload, ok := ev.payload.(domain.StatusChangedPayload)
	if !ok {
		t.Fatalf("payload type=%T, want StatusChangedPayload", ev.payload)
	}
	if payload.TaskID != id || payload.Status != domain.StatusInProgress {
		t.Errorf("payload=%+v, want {TaskID: %v, Status: %v}", payload, id, domain.StatusInProgress)
	}
}
