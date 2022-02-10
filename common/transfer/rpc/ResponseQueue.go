package rpc

import (
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/transfer/rpc/jsonrpc"
	"returntypes-langserver/common/utils"
	"sync"
)

// A thread-safe random access queue holding responses.
type responseQueue struct {
	responses []jsonrpc.Response
	mutex     sync.Mutex
	revision  utils.Revision
	closed    bool
}

// Appends a new response to the queue
func (q *responseQueue) Append(response jsonrpc.Response) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	if !q.closed {
		q.responses = append(q.responses, response)
		if q.revision != nil {
			q.revision.SetOutdated()
		}
		q.revision = utils.NewRevision()
	}
}

// Picks a response with the given id. Blocks, until a response is found.
func (q *responseQueue) PickResponseWithId(id int) (jsonrpc.Response, errors.Error) {
	for !q.closed {
		response, currentRevision := q.pickResponseWithId(id)
		if currentRevision != nil {
			// pickResponseWithId returns the current revision (at the time picking) if a response was not found
			// so wait until it is outdated.
			currentRevision.WaitUntilOutdated()
		} else if !q.closed {
			return response, nil
		}
	}
	return jsonrpc.Response{}, errors.New("Error", "No response found due to queue being closed")
}

func (q *responseQueue) pickResponseWithId(id int) (jsonrpc.Response, utils.Revision) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	if !q.closed {
		if q.responses != nil {
			// search for a response with the given id to pick and remove from the slice
			for i, response := range q.responses {
				if responseId, ok := response.Id.(float64); ok && int(responseId) == id {
					if len(q.responses) > i+1 {
						copy(q.responses[i:], q.responses[i+1:])
					}
					q.responses = q.responses[:len(q.responses)-1]
					return response, nil
				}
			}
		}
		if q.revision == nil {
			q.revision = utils.NewRevision()
		}
		// return current revision to block until an update is pushed
		currentRevision := q.revision
		return jsonrpc.Response{}, currentRevision
	}
	return jsonrpc.Response{}, nil
}

// Closes the response queue and releases all waiting go routines
func (q *responseQueue) Close() {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.closed {
		return
	}

	q.closed = true
	q.responses = nil
	if q.revision != nil {
		q.revision.SetOutdated()
	}
}

// Reopens the response queue after it was closed.
func (q *responseQueue) Reopen() {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if !q.closed {
		return
	}

	q.closed = false
}
