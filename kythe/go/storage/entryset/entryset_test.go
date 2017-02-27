/*
 * Copyright 2017 Google Inc. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package entryset

import (
	"log"
	"reflect"
	"sort"
	"strings"
	"testing"

	"kythe.io/kythe/go/util/kytheuri"

	spb "kythe.io/kythe/proto/storage_proto"
)

func TestCanon(t *testing.T) {
	s := New()
	if !s.canon {
		t.Error("A New set should be canonical")
	}

	s.Add(&spb.Entry{
		Source:   &spb.VName{Corpus: "foo"},
		Target:   &spb.VName{Corpus: "bar"},
		EdgeKind: "baz",
	})
	if s.canon {
		t.Error("A modified set should not be canonical")
	}

	s.Canonicalize()
	if !s.canon {
		t.Error("A canonicalized set should be canonical")
	}
}

func TestSorting(t *testing.T) {
	input := []string{"some", "of", "what", "a", "fool", "thinks", "often", "remains"}

	result := make([]string, len(input))
	copy(result, input)
	rev := sortInverse(sort.StringSlice(result))
	t.Log("Checking order...")
	for i, s := range result {
		t.Logf("result[%d] = %q", i, s)
		if i+1 < len(result) && s >= result[i+1] {
			t.Errorf("At offset %d: %q >= %q", i, s, result[i+1])
		}
	}
	t.Log("Checking reverse permutation...")
	for i, p := range rev {
		t.Logf("result[%d] = %q", p, result[p])
		if result[p] != input[i] {
			t.Errorf("At offset %d: got %q, want %q", i, result[p], input[i])
		}
	}
}

func F(ticket, name, value string) *spb.Entry {
	v, err := kytheuri.ToVName(ticket)
	if err != nil {
		log.Fatalf("Invalid ticket %q: %v", ticket, err)
	}
	return &spb.Entry{
		Source:    v,
		FactName:  name,
		FactValue: []byte(value),
	}
}

func E(source, target, kind string) *spb.Entry {
	src, err := kytheuri.ToVName(source)
	if err != nil {
		log.Fatalf("Invalid source ticket %q: %v", source, err)
	}
	tgt, err := kytheuri.ToVName(target)
	if err != nil {
		log.Fatalf("Invalid target ticket %q: %v", target, err)
	}
	return &spb.Entry{
		Source:   src,
		Target:   tgt,
		EdgeKind: kind,
	}
}

func estr(e *spb.Entry) [13]string {
	s, t := e.Source, e.Target
	return [...]string{
		s.Corpus, s.Language, s.Path, s.Root, s.Signature,
		e.EdgeKind,
		t.GetCorpus(), t.GetLanguage(), t.GetPath(), t.GetRoot(), t.GetSignature(),
		e.FactName, string(e.FactValue),
	}
}

func ecompare(e1, e2 *spb.Entry) int {
	s1 := estr(e1)
	s2 := estr(e2)
	for i, a := range s1 {
		if c := strings.Compare(a, s2[i]); c != 0 {
			return c
		}
	}
	return 0
}

type byEntryOrder []*spb.Entry

func (e byEntryOrder) Len() int           { return len(e) }
func (e byEntryOrder) Less(i, j int) bool { return ecompare(e[i], e[j]) < 0 }
func (e byEntryOrder) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }

func (e byEntryOrder) Sorted() []*spb.Entry {
	cp := make([]*spb.Entry, len(e))
	copy(cp, []*spb.Entry(e))
	sort.Sort(byEntryOrder(cp))
	return cp
}

var testEntries = []*spb.Entry{
	// These entries are intentionally not in canonical order.

	E("//bravo?root=R?path=P#sig", "//alpha?path=p/r", "validates"),
	F("//alpha?path=p/q", "node/kind", "file"),
	F("//alpha?path=p/r", "node/kind", "file"),
	E("//alpha?path=p/q", "//alpha?path=p/r", "includes"),
	F("//bravo?path=P?root=R#sig", "loc/start", "122"),
	E("//gamma#blah", "//gamma#blah", "selfloop"),
	E("//a#first", "//b#second", "outbound"),
	E("//b#second", "//a#first", "return"),
	F("//b#second", "eats", "cabbage"),
}

func testSet(t *testing.T) *Set {
	s := New()
	for _, entry := range testEntries {
		if err := s.Add(entry); err != nil {
			t.Fatalf("Error adding entry: %v", err)
		}
	}
	return s
}

func TestVisit(t *testing.T) {
	s := testSet(t)

	want := byEntryOrder(testEntries).Sorted()
	var got []*spb.Entry
	s.Visit(func(e *spb.Entry) bool {
		got = append(got, e)
		return true
	})
	sort.Sort(byEntryOrder(got))

	// Verify that we got all the entries and no extras.
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Visit order differs.\n--- Got:\n%+v\n--- Want:\n%+v\n", got, want)
	}
}

func TestRoundTrip(t *testing.T) {
	before := testSet(t).Canonicalize()

	var beforeEntries []*spb.Entry
	before.Visit(func(e *spb.Entry) bool {
		beforeEntries = append(beforeEntries, e)
		return true
	})

	after, err := Decode(before.Encode())
	if err != nil {
		t.Errorf("Decoding from protobuf failed: %v", err)
	}

	var afterEntries []*spb.Entry
	after.Visit(func(e *spb.Entry) bool {
		afterEntries = append(afterEntries, e)
		return true
	})

	// Because the input was canonical before encoding, the decoded result
	// should come back in exactly the same (canonical) order.

	if !reflect.DeepEqual(beforeEntries, afterEntries) {
		t.Errorf("Round-trip failed.\n--- Got:\n%+v\n--- Want:\n%+v\n", beforeEntries, afterEntries)
	}
}