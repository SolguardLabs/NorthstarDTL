package settlement

import (
	"sort"

	"github.com/solguardlabs/northstardtl/src/domain"
)

type Queue struct {
	order   []domain.TicketID
	tickets map[domain.TicketID]domain.Ticket
}

func NewQueue() *Queue {
	return &Queue{tickets: make(map[domain.TicketID]domain.Ticket)}
}

func (queue *Queue) Add(ticket domain.Ticket) error {
	if _, ok := queue.tickets[ticket.ID]; ok {
		return domain.Duplicate("ticket %s already exists", ticket.ID)
	}
	queue.order = append(queue.order, ticket.ID)
	queue.tickets[ticket.ID] = ticket
	return nil
}

func (queue *Queue) Get(id domain.TicketID) (domain.Ticket, bool) {
	ticket, ok := queue.tickets[id]
	return ticket, ok
}

func (queue *Queue) Update(ticket domain.Ticket) error {
	if _, ok := queue.tickets[ticket.ID]; !ok {
		return domain.NewError(domain.ErrTicketNotFound, "ticket %s not found", ticket.ID)
	}
	queue.tickets[ticket.ID] = ticket
	return nil
}

func (queue *Queue) Ready(now int64, count int) []domain.Ticket {
	queue.refresh(now)
	tickets := make([]domain.Ticket, 0)
	for _, id := range queue.order {
		ticket := queue.tickets[id]
		if ticket.IsReady(now) {
			tickets = append(tickets, ticket)
		}
	}
	sort.SliceStable(tickets, func(i, j int) bool {
		if tickets[i].QueueScore != tickets[j].QueueScore {
			return tickets[i].QueueScore > tickets[j].QueueScore
		}
		if tickets[i].ReadyEpoch != tickets[j].ReadyEpoch {
			return tickets[i].ReadyEpoch < tickets[j].ReadyEpoch
		}
		return tickets[i].ID < tickets[j].ID
	})
	if count > 0 && count < len(tickets) {
		return tickets[:count]
	}
	return tickets
}

func (queue *Queue) Pending() []domain.Ticket {
	out := make([]domain.Ticket, 0, len(queue.tickets))
	for _, id := range queue.order {
		ticket := queue.tickets[id]
		if ticket.Status != domain.TicketExecuted && ticket.Status != domain.TicketRejected {
			out = append(out, ticket)
		}
	}
	return out
}

func (queue *Queue) All() []domain.Ticket {
	out := make([]domain.Ticket, 0, len(queue.tickets))
	for _, id := range queue.order {
		out = append(out, queue.tickets[id])
	}
	return out
}

func (queue *Queue) Snapshot() []domain.QueueSnapshot {
	pending := queue.Pending()
	sort.SliceStable(pending, func(i, j int) bool {
		if pending[i].QueueScore != pending[j].QueueScore {
			return pending[i].QueueScore > pending[j].QueueScore
		}
		return pending[i].ID < pending[j].ID
	})
	out := make([]domain.QueueSnapshot, len(pending))
	for index, ticket := range pending {
		out[index] = domain.NewQueueSnapshot(ticket)
	}
	return out
}

func (queue *Queue) refresh(now int64) {
	for _, id := range queue.order {
		ticket := queue.tickets[id]
		updated := ticket.MarkReady(now)
		if updated.Status != ticket.Status {
			queue.tickets[id] = updated
		}
	}
}
