package ccv2_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("OrganizationQuota", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("GetOrganizationQuota", func() {

		When("getting the organization quota does not return an error", func() {
			BeforeEach(func() {
				response := `{
				"metadata": {
					"guid": "some-org-quota-guid"
				},
				"entity": {
					"name": "some-org-quota"
				}
			}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/quota_definitions/some-org-quota-guid"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					),
				)
			})

			It("returns the organization quota", func() {
				orgQuota, warnings, err := client.GetOrganizationQuota("some-org-quota-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(Equal(Warnings{"warning-1"}))
				Expect(orgQuota).To(Equal(OrganizationQuota{
					GUID: "some-org-quota-guid",
					Name: "some-org-quota",
				}))
			})
		})

		When("the organization quota returns an error", func() {
			BeforeEach(func() {
				response := `{
				  "description": "Quota Definition could not be found: some-org-quota-guid",
				  "error_code": "CF-QuotaDefinitionNotFound",
				  "code": 240001
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/quota_definitions/some-org-quota-guid"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					),
				)
			})

			It("returns the error", func() {
				_, warnings, err := client.GetOrganizationQuota("some-org-quota-guid")
				Expect(err).To(MatchError(ccerror.ResourceNotFoundError{
					Message: "Quota Definition could not be found: some-org-quota-guid",
				}))
				Expect(warnings).To(Equal(Warnings{"warning-1"}))
			})
		})

	})

	FDescribe("GetOrganizationQuotaByName", func() {
		Context("when quota exists", func() {
			BeforeEach(func() {
				response := `{
				"metadata": {
					"guid": "some-org-quota-guid"
				},
				"entity": {
					"name": "some-org-quota-name"
				}
			}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/quota_definitions?q=name:some-org-quota-name"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					),
				)
			})

			It("returns the organization quota", func() {
				orgQuota, warnings, err := client.GetOrganizationQuotaByName("some-org-quota-name")
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(Equal(Warnings{"warning-1"}))
				Expect(orgQuota).To(Equal(OrganizationQuota{
					GUID: "some-org-quota-guid",
					Name: "some-org-quota-name",
				}))

			})
		})

		Context("when quota doesn't exist", func() {
			BeforeEach(func() {
				response := `{
				"metadata": {
					"guid": "some-org-quota-guid"
				},
				"entity": {
					"name": "some-bogus-name"
				}
			}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/quota_definitions?q=name:some-org-quota-name"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					),
				)
			})

			It("returns warnings and errors", func() {
				_, warnings, err := client.GetOrganizationQuotaByName("some-bogus-name")
				Expect(err).To(MatchError("some-error"))
				Expect(warnings).To(Equal(Warnings{"warning-1"}))
			})
		})
	})
})
