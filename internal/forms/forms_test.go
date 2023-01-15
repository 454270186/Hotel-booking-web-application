package forms

import (
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestForm_Valid(t *testing.T) {
	r := httptest.NewRequest("POST", "/some-url", nil)
	form := New(r.PostForm)

	if !form.Valid() {
		t.Error("empty form should be valid")
	}

}

func TestForm_Required(t *testing.T) {
	r := httptest.NewRequest("POST", "/some-url", nil)
	form := New(r.PostForm)

	form.Required("a", "b", "c")
	if form.Valid() {
		t.Error("This test should fail")
	}

	postData := url.Values{}
	postData.Add("a", "1")
	postData.Add("b", "2")
	postData.Add("c", "3")

	// reset Request
	r = httptest.NewRequest("POST", "/some-url", nil)
	r.PostForm = postData

	form = New(r.PostForm)

	form.Required("a", "b", "c")
	if !form.Valid() {
		t.Error("This test should not fail")
	}
}

func TestForm_Has(t *testing.T) {
	postData := url.Values{}
	postData.Add("a", "1")
	postData.Add("b", "")

	form := New(postData)
	if form.Has("b") {
		t.Error("Key-b is empty, should return false")
	}
	if !form.Has("a") {
		t.Error("Key-a is not empty, should return true")
	}
}

func TestForm_MinLength(t *testing.T) {
	postData := url.Values{}
	postData.Add("a", "123")
	postData.Add("b", "123")

	form1 := New(postData)
	form1.MinLength("a", 3)
	if !form1.Valid() {
		t.Error("length of Value-a is equal to 3, this test should not fail")
	}

	form2 := New(postData)
	form2.MinLength("b", 4)
	if form2.Valid() {
		t.Error("length of Value-b is less than 4, this test should fail")
	}

}

func TestForm_IsEmail(t *testing.T) {
	r := httptest.NewRequest("POST", "/", nil)
	postData := url.Values{}
	postData.Add("Good-email", "sc21ey@leeds.ac.uk")
	postData.Add("Bad-email", "123456")
	r.PostForm = postData

	form := New(r.PostForm)
	form.IsEmail("Good-email")
	if !form.Valid() {
		t.Error("this email is valid, should not fail")
	}
	form.IsEmail("Bad-email")
	if form.Valid() {
		t.Error("this email is invalid, should fail")
	}

}
