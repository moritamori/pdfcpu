/*
Copyright 2019 The pdf Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package test

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
)

func prepareForAttachmentTest(t *testing.T) error {
	t.Helper()
	for _, fileName := range []string{"go.pdf", "golang.pdf", "T4.pdf", "go-lecture.pdf"} {
		inFile := filepath.Join(inDir, fileName)
		outFile := filepath.Join(outDir, fileName)
		if err := copyFile(t, inFile, outFile); err != nil {
			return err
		}
	}
	return copyFile(t, filepath.Join(resDir, "test.wav"), filepath.Join(outDir, "test.wav"))
}

func listAttachments(t *testing.T, msg, fileName string, want int) []string {
	t.Helper()

	list, err := api.ListAttachmentsFile(fileName, nil)
	if err != nil {
		t.Fatalf("%s list attachments: %v\n", msg, err)
	}

	// # of attachments must be want
	if len(list) != want {
		t.Fatalf("%s: list attachments %s: want %d got %d\n", msg, fileName, want, len(list))
	}
	return list
}

func TestAttachments(t *testing.T) {
	msg := "testAttachments"

	if err := prepareForAttachmentTest(t); err != nil {
		t.Fatalf("%s prepare for attachments: %v\n", msg, err)
	}

	fileName := filepath.Join(outDir, "go.pdf")

	// # of attachments must be 0
	listAttachments(t, msg, fileName, 0)

	// attach add 4 files
	files := []string{
		outDir + "/golang.pdf",
		outDir + "/T4.pdf",
		outDir + "/go-lecture.pdf",
		outDir + "/test.wav"}

	if err := api.AddAttachmentsFile(fileName, "", files, false, nil); err != nil {
		t.Fatalf("%s add attachments: %v\n", msg, err)
	}
	list := listAttachments(t, msg, fileName, 4)
	for _, s := range list {
		t.Log(s)
	}

	// Extract all attachments.
	if err := api.ExtractAttachmentsFile(fileName, outDir, nil, nil); err != nil {
		t.Fatalf("%s extract all attachments: %v\n", msg, err)
	}

	// Extract 1 attachment.
	if err := api.ExtractAttachmentsFile(fileName, outDir, []string{"golang.pdf"}, nil); err != nil {
		t.Fatalf("%s extract one attachment: %v\n", msg, err)
	}

	// Remove 1 attachment.
	if err := api.RemoveAttachmentsFile(fileName, "", []string{"golang.pdf"}, nil); err != nil {
		t.Fatalf("%s remove one attachment: %v\n", msg, err)
	}
	listAttachments(t, msg, fileName, 3)

	// Remove all attachments.
	if err := api.RemoveAttachmentsFile(fileName, "", nil, nil); err != nil {
		t.Fatalf("%s remove all attachments: %v\n", msg, err)
	}
	listAttachments(t, msg, fileName, 0)

	// Validate the processed file.
	if err := api.ValidateFile(fileName, nil); err != nil {
		t.Fatalf("%s: validate: %v\n", msg, err)
	}
}

func TestAttachmentUsingReader(t *testing.T) {
	msg := "TestAttachmentUsingStringReader"

	file := "go.pdf"
	inFile := filepath.Join(inDir, file)
	outFile := filepath.Join(outDir, file)
	if err := copyFile(t, inFile, outFile); err != nil {
		t.Fatalf("%s copyFile: %v\n", msg, err)
	}

	// Create a context.
	ctx, err := api.ReadContextFile(outFile)
	if err != nil {
		t.Fatalf("%s readContext: %v\n", msg, err)
	}

	// List attachments.
	if aa, err := ctx.ListAttachments(); err != nil || len(aa) > 0 {
		t.Fatalf("%s listAttachments: %v\n", msg, err)
	}

	// Add attachment.
	id := "attachment1"
	want := "12345"
	now := time.Now()
	a := pdfcpu.NewAttachment(strings.NewReader(want), id, "description", &now)
	useCollection := false
	if err = ctx.AddAttachment(*a, useCollection); err != nil {
		t.Fatalf("%s addAttachment: %v\n", msg, err)
	}

	// List attachments.
	aa, err := ctx.ListAttachments()
	if err != nil || len(aa) != 1 || aa[0].ID != id {
		t.Fatalf("%s listAttachments: %v\n", msg, err)
	}

	// Extract attachment.
	aa, err = ctx.ExtractAttachments([]string{id})
	if err != nil {
		t.Fatalf("%s extractAttachment: %v\n", msg, err)
	}
	if len(aa) != 1 { // || aa == nil
		t.Fatalf("%s extractAttachment: attachment %s not found\n", msg, id)
	}
	gotBytes, err := aa[0].Bytes()
	if err != nil {
		t.Fatalf("%s extractAttachment: attachment %s no data available\n", msg, id)
	}
	got := string(gotBytes)
	if got != want {
		t.Fatalf("%s\ngot:%s\nwant:%s", msg, got, want)
	}

	if err := api.WriteContextFile(ctx, outFile); err != nil {
		t.Fatalf("%s writeContext: \n", msg, err)
	}

	// Optional processing of attachment bytes.

	// Remove attachment.
	ok, err := ctx.RemoveAttachments([]string{id})
	if err != nil {
		t.Fatalf("%s removeAttachment: %v\n", msg, err)
	}
	if !ok {
		t.Fatalf("%s removeAttachment: attachment %s not found\n", msg, id)
	}

	// List attachment.
	if aa, err = ctx.ListAttachments(); err != nil || len(aa) > 0 {
		t.Fatalf("%s listAttachments: %v\n", msg, err)
	}
}
