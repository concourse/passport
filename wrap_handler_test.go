package auth_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/concourse/atc/auth"
	"github.com/concourse/atc/auth/fakes"
)

var _ = Describe("WrapHandler", func() {
	var (
		fakeValidator         *fakes.FakeValidator
		fakeUserContextReader *fakes.FakeUserContextReader

		server *httptest.Server
		client *http.Client

		authenticated <-chan bool
		teamNameChan  <-chan string
		teamIDChan    <-chan int
		isAdminChan   <-chan bool
		foundChan     <-chan bool
	)

	BeforeEach(func() {
		fakeValidator = new(fakes.FakeValidator)
		fakeUserContextReader = new(fakes.FakeUserContextReader)

		a := make(chan bool, 1)
		tn := make(chan string, 1)
		ti := make(chan int, 1)
		ia := make(chan bool, 1)
		f := make(chan bool, 1)
		authenticated = a
		teamNameChan = tn
		teamIDChan = ti
		isAdminChan = ia
		foundChan = f
		simpleHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			a <- auth.IsAuthenticated(r)
			teamName, teamID, isAdmin, found := auth.GetTeam(r)
			f <- found
			tn <- teamName
			ti <- teamID
			ia <- isAdmin
		})

		server = httptest.NewServer(auth.WrapHandler(
			simpleHandler,
			fakeValidator,
			fakeUserContextReader,
		))

		client = &http.Client{
			Transport: &http.Transport{},
		}
	})

	Context("when a request is made", func() {
		var request *http.Request
		var response *http.Response

		BeforeEach(func() {
			var err error

			request, err = http.NewRequest("GET", server.URL, bytes.NewBufferString("hello"))
			Expect(err).NotTo(HaveOccurred())
		})

		JustBeforeEach(func() {
			var err error

			response, err = client.Do(request)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when the validator returns true", func() {
			BeforeEach(func() {
				fakeValidator.IsAuthenticatedReturns(true)
			})

			It("handles the request with the request authenticated", func() {
				Expect(<-authenticated).To(BeTrue())
			})
		})

		Context("when the validator returns false", func() {
			BeforeEach(func() {
				fakeValidator.IsAuthenticatedReturns(false)
			})

			It("handles the request with the request authenticated", func() {
				Expect(<-authenticated).To(BeFalse())
			})
		})

		Context("when the userContextReader finds team information", func() {
			BeforeEach(func() {
				fakeUserContextReader.GetTeamReturns("some-team", 9, true, true)
			})

			It("passes the team information along in the request object", func() {
				Expect(<-foundChan).To(BeTrue())
				Expect(<-teamNameChan).To(Equal("some-team"))
				Expect(<-teamIDChan).To(Equal(9))
				Expect(<-isAdminChan).To(BeTrue())
			})
		})

		Context("when the userContextReader does not find team information", func() {
			BeforeEach(func() {
				fakeUserContextReader.GetTeamReturns("", 0, false, false)
			})

			It("does not pass team information along in the request object", func() {
				Expect(<-foundChan).To(BeFalse())
			})
		})
	})
})
