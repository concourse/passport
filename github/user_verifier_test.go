package github_test

import (
	"errors"
	"net/http"

	. "github.com/concourse/atc/auth/github"
	"github.com/concourse/atc/auth/github/fakes"
	"github.com/pivotal-golang/lager/lagertest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UserVerifier", func() {
	var (
		fakeClient *fakes.FakeClient

		verifier Verifier
	)

	BeforeEach(func() {
		fakeClient = new(fakes.FakeClient)

		verifier = NewUserVerifier([]string{"some-user", "some-other-user"}, fakeClient)
	})

	Describe("Verify", func() {
		var (
			httpClient *http.Client

			verified  bool
			verifyErr error
		)

		BeforeEach(func() {
			httpClient = &http.Client{}
		})

		JustBeforeEach(func() {
			verified, verifyErr = verifier.Verify(lagertest.NewTestLogger("test"), httpClient)
		})

		Context("when the client returns the current user", func() {
			Context("when the user is permitted", func() {
				BeforeEach(func() {
					fakeClient.CurrentUserReturns("some-user", nil)
				})

				It("succeeds", func() {
					Expect(verifyErr).ToNot(HaveOccurred())
				})

				It("returns true", func() {
					Expect(verified).To(BeTrue())
				})
			})

			Context("when the user is not permitted", func() {
				BeforeEach(func() {
					fakeClient.CurrentUserReturns("bogus-user", nil)
				})

				It("succeeds", func() {
					Expect(verifyErr).ToNot(HaveOccurred())
				})

				It("returns false", func() {
					Expect(verified).To(BeFalse())
				})
			})
		})

		Context("when the client fails", func() {
			disaster := errors.New("nope")

			BeforeEach(func() {
				fakeClient.CurrentUserReturns("", disaster)
			})

			It("returns the error", func() {
				Expect(verifyErr).To(Equal(disaster))
			})
		})
	})
})
