/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package model

import (
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// Tests loading annotations from file, writing back out and reloading.
func TestReadWriteAnnotations(t *testing.T) {
	f, err := os.Open(`testdata/OoPdfFormExample.pdf`)
	require.NoError(t, err)
	defer f.Close()

	reader, err := NewPdfReaderLazy(f)
	require.NoError(t, err)

	checkAnnots := func(reader *PdfReader, formExpected bool) {
		// Check Acroform and fields.
		if formExpected {
			require.NotNil(t, reader.AcroForm)
			fields := reader.AcroForm.AllFields()
			require.Len(t, fields, 17)
			require.Nil(t, fields[0].Parent)
		} else {
			require.Nil(t, reader.AcroForm)
		}

		// Check annotations.
		numPages, err := reader.GetNumPages()
		require.NoError(t, err)
		require.Equal(t, 1, numPages)

		page, err := reader.GetPage(1)
		require.NoError(t, err)

		require.NotNil(t, page.Annots)
		annots, err := page.GetAnnotations()
		require.NoError(t, err)
		require.Len(t, annots, 17)

		wa, ok := annots[0].GetContext().(*PdfAnnotationWidget)
		require.True(t, ok)
		if formExpected {
			require.NotNil(t, wa.parent)
			require.NotNil(t, wa.Parent)
		} else {
			require.Nil(t, wa.parent)
			require.Nil(t, wa.Parent)
		}
	}
	checkAnnots(reader, true)

	// Write out and reload. With the AcroForm in place.
	{
		w := NewPdfWriter()
		page, err := reader.GetPage(1)
		require.NoError(t, err)
		err = w.AddPage(page)
		require.NoError(t, err)
		err = w.SetForms(reader.AcroForm)
		require.NoError(t, err)

		var buf bytes.Buffer
		err = w.Write(&buf)
		require.NoError(t, err)

		// For debugging:
		// ioutil.WriteFile("/tmp/test_read_write_annots_withacro.pdf", buf.Bytes(), 0644)

		bufReader := bytes.NewReader(buf.Bytes())
		reader, err = NewPdfReaderLazy(bufReader)
		require.NoError(t, err)

		checkAnnots(reader, true)
	}

	// Write out and reload without setting the AcroForm.
	{
		w := NewPdfWriter()
		page, err := reader.GetPage(1)
		require.NoError(t, err)
		err = w.AddPage(page)
		require.NoError(t, err)

		var buf bytes.Buffer
		err = w.Write(&buf)
		require.NoError(t, err)

		// For debugging:
		// ioutil.WriteFile("/tmp/test_read_write_annots_noacro.pdf", buf.Bytes(), 0644)

		bufReader := bytes.NewReader(buf.Bytes())
		reader, err = NewPdfReaderLazy(bufReader)
		require.NoError(t, err)

		checkAnnots(reader, false)
	}
}

// erroringWriter errors on write for testing.
type erroringWriter struct{}

// Write returns error on writing.
func (w *erroringWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("error")
}

// TestWriterErrorHandling tests error handling of the writer.
// https://github.com/mgmeyers/unipdf/issues/316
func TestWriterErrorHandling(t *testing.T) {
	w := NewPdfWriter()
	page := NewPdfPage()
	err := w.AddPage(page)
	require.NoError(t, err)

	// Errors in writing should be passed up.
	out := erroringWriter{}
	err = w.Write(&out)
	require.Error(t, err)
}
