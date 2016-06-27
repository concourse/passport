package auth_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/concourse/atc/auth"
	"github.com/concourse/atc/auth/authfakes"
)

var _ = Describe("ValidatorBasket", func() {
	var (
		fakeValidator1  *authfakes.FakeValidator
		fakeValidator2  *authfakes.FakeValidator
		validatorBasket auth.ValidatorBasket
	)

	BeforeEach(func() {
		fakeValidator1 = new(authfakes.FakeValidator)
		fakeValidator2 = new(authfakes.FakeValidator)

		validatorBasket = auth.ValidatorBasket{
			fakeValidator1,
			fakeValidator2,
		}
	})

	It("fails to authenticate if none of the embedded validators return true", func() {
		fakeValidator1.IsAuthenticatedReturns(false)
		fakeValidator2.IsAuthenticatedReturns(false)

		result := validatorBasket.IsAuthenticated(&http.Request{})
		Expect(result).To(BeFalse())
	})

	It("authenticates if any of the embedded validators return true", func() {
		fakeValidator1.IsAuthenticatedReturns(false)
		fakeValidator2.IsAuthenticatedReturns(true)

		result := validatorBasket.IsAuthenticated(&http.Request{})
		Expect(result).To(BeTrue())

		fakeValidator1.IsAuthenticatedReturns(true)
		fakeValidator2.IsAuthenticatedReturns(false)

		result = validatorBasket.IsAuthenticated(&http.Request{})
		Expect(result).To(BeTrue())
	})
})
