// Copyright (C) 2025 Storj Labs, Inc.
// See LICENSE for copying information.

package cohorts

import (
	"sync"

	"storj.io/common/pb"
)

// Matcher is a type that takes in some requirements and node tags and can be told when a node has
// succeeded and it will report if the set of successful nodes meets the requirements.
type Matcher struct {
	child matcherDispatch
	tags  []map[string]string

	mu        sync.Mutex
	successes bitset
}

// NewMatcher creates a new matcher with the given node tags and requirements.
func NewMatcher(reqs *pb.CohortRequirements, totalPieces int) *Matcher {
	return &Matcher{
		child: newMatcherDispatch(reqs),
		tags:  make([]map[string]string, totalPieces),
	}
}

// Increment marks a node as successful.
func (m *Matcher) Increment(tags map[string]string, node byte) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.successes.has(node) {
		m.successes.set(node)
		m.tags[node] = tags
	}
	return m.child.valid(&m.successes, m.tags)
}

//
// matcherDispatch is like an interface but avoids virtual dispatch so the compiler can avoid heap
// allocating things.
//

type matcherDispatch struct {
	and  *matcherAnd
	lit  *matcherLiteral
	with *matcherWithhold
}

func newMatcherDispatch(reqs *pb.CohortRequirements) matcherDispatch {
	if reqs == nil {
		return matcherDispatch{}
	}

	switch req := reqs.Requirement.(type) {
	case *pb.CohortRequirements_And_:
		children := make([]matcherDispatch, len(req.And.Requirements))
		for i, child := range req.And.Requirements {
			children[i] = newMatcherDispatch(child)
		}
		return newMatcherAnd(children...)

	case *pb.CohortRequirements_Literal_:
		return newMatcherLiteral(int(req.Literal.Value))

	case *pb.CohortRequirements_Withhold_:
		return newMatcherWithhold(
			req.Withhold.TagKey,
			int(req.Withhold.Amount),
			newMatcherDispatch(req.Withhold.Child),
		)

	default:
		return matcherDispatch{}
	}
}

func (m matcherDispatch) valid(successes *bitset, tags []map[string]string) bool {
	switch {
	case m.and != nil:
		return m.and.valid(successes, tags)
	case m.lit != nil:
		return m.lit.valid(successes, tags)
	case m.with != nil:
		return m.with.valid(successes, tags)
	default:
		return true
	}
}

func (m matcherDispatch) min() int {
	switch {
	case m.and != nil:
		return m.and.min()
	case m.lit != nil:
		return m.lit.min()
	case m.with != nil:
		return m.with.min()
	default:
		return 0
	}
}

//
// matcherLiteral is a matcher that checks if the number of successes is greater than or equal to a
// literal value.
//

type matcherLiteral struct {
	value int
}

func newMatcherLiteral(value int) matcherDispatch {
	return matcherDispatch{lit: &matcherLiteral{
		value: value,
	}}
}

func (m *matcherLiteral) valid(successes *bitset, tags []map[string]string) bool {
	return successes.count() >= m.value
}

func (m *matcherLiteral) min() int {
	return m.value
}

//
// matcherAnd is a matcher that requires both matchers to be valid.
//

type matcherAnd struct {
	children []matcherDispatch
	minimum  int
}

func newMatcherAnd(children ...matcherDispatch) matcherDispatch {
	minimum := 0
	for _, child := range children {
		minimum = max(minimum, child.min())
	}
	return matcherDispatch{and: &matcherAnd{
		children: children,
		minimum:  minimum,
	}}
}

func (m *matcherAnd) valid(successes *bitset, tags []map[string]string) bool {
	if successes.count() < m.min() {
		return false
	}
	for _, child := range m.children {
		if !child.valid(successes, tags) {
			return false
		}
	}
	return true
}

func (m *matcherAnd) min() int {
	return m.minimum
}

//
// matcherWithhold is a matcher that removes the top N values from the successes bitset before
// passing it down to the child matcher.
//

type matcherWithhold struct {
	key    string // which tag key to use
	amount int    // how many top values to remove
	child  matcherDispatch

	// reused maps for withhold processing
	counts map[string]int
	topn   map[string]int
}

func newMatcherWithhold(key string, amount int, child matcherDispatch) matcherDispatch {
	return matcherDispatch{with: &matcherWithhold{
		key:    key,
		amount: amount,
		child:  child,
	}}
}

func (m *matcherWithhold) valid(successes *bitset, tags []map[string]string) bool {
	if successes.count() < m.min() {
		return false
	}

	clear(m.counts)
	if m.counts == nil {
		m.counts = make(map[string]int)
	}

	clear(m.topn)
	if m.topn == nil {
		m.topn = make(map[string]int, m.amount)
	}

	for i := range successes.iter {
		m.counts[tags[i][m.key]]++
	}

	for value, count := range m.counts {
		if value == "" {
			continue
		}

		if len(m.topn) < m.amount {
			m.topn[value] = count
			continue
		}

		for checkValue, checkCount := range m.topn {
			if count > checkCount {
				delete(m.topn, checkValue)
				m.topn[value] = count
				break
			}
		}
	}

	childSuccesses := *successes
	for i := range childSuccesses.iter {
		if _, ok := m.topn[tags[i][m.key]]; ok {
			childSuccesses.unset(i)
		}
	}

	return m.child.valid(&childSuccesses, tags)
}

func (m *matcherWithhold) min() int {
	return m.child.min() + m.amount // each failure must remove at least one
}
