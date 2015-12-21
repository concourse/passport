package github_test

import (
	"errors"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager/lagertest"

	afakes "github.com/concourse/atc/auth/fakes"

	. "github.com/concourse/atc/auth/github"
)

var _ = Describe("VerifierBasket", func() {
	var (
		fakeVerifier1 *afakes.FakeVerifier
		fakeVerifier2 *afakes.FakeVerifier

		httpClient     *http.Client
		verifierBasket Verifier
	)

	BeforeEach(func() {

		fakeVerifier1 = new(afakes.FakeVerifier)
		fakeVerifier2 = new(afakes.FakeVerifier)

		httpClient = &http.Client{}
		verifierBasket = NewVerifierBasket(fakeVerifier1, fakeVerifier2)
	})

	It("fails to verify if none of the passed in verifiers return true", func() {
		fakeVerifier1.VerifyReturns(false, nil)
		fakeVerifier2.VerifyReturns(false, nil)

		result, err := verifierBasket.Verify(lagertest.NewTestLogger("test"), httpClient)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(BeFalse())
	})

	It("verifies if any of the embedded verifiers return true", func() {
		fakeVerifier1.VerifyReturns(false, nil)
		fakeVerifier2.VerifyReturns(true, nil)

		result, err := verifierBasket.Verify(lagertest.NewTestLogger("test"), httpClient)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(BeTrue())

		fakeVerifier1.VerifyReturns(true, nil)
		fakeVerifier2.VerifyReturns(false, nil)

		result, err = verifierBasket.Verify(lagertest.NewTestLogger("test"), httpClient)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(BeTrue())
	})

	It("errors if all of the embedded verifiers error", func() {
		fakeVerifier1.VerifyReturns(false, errors.New("first error"))
		fakeVerifier2.VerifyReturns(false, errors.New("second error"))

		_, err := verifierBasket.Verify(lagertest.NewTestLogger("test"), httpClient)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("first error"))
		Expect(err.Error()).To(ContainSubstring("second error"))
	})

	It("errors if no verifiers return true and at least one errors", func() {
		fakeVerifier1.VerifyReturns(false, errors.New("first error"))
		fakeVerifier2.VerifyReturns(false, nil)

		_, err := verifierBasket.Verify(lagertest.NewTestLogger("test"), httpClient)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("first error"))
	})

	It("does not error if at least one verifier returns true", func() {
		fakeVerifier1.VerifyReturns(false, errors.New("first error"))
		fakeVerifier2.VerifyReturns(true, nil)

		result, err := verifierBasket.Verify(lagertest.NewTestLogger("test"), httpClient)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(BeTrue())
	})
})
