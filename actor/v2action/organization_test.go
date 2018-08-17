package v2action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Org Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil)
	})

	Describe("GetOrganization", func() {
		var (
			org      Organization
			warnings Warnings
			err      error
		)

		JustBeforeEach(func() {
			org, warnings, err = actor.GetOrganization("some-org-guid")
		})

		When("the org exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationReturns(
					ccv2.Organization{
						GUID:                "some-org-guid",
						Name:                "some-org",
						QuotaDefinitionGUID: "some-quota-definition-guid",
					},
					ccv2.Warnings{"warning-1", "warning-2"},
					nil)
			})

			It("returns the org and all warnings", func() {
				Expect(err).ToNot(HaveOccurred())

				Expect(org.GUID).To(Equal("some-org-guid"))
				Expect(org.Name).To(Equal("some-org"))
				Expect(org.QuotaDefinitionGUID).To(Equal("some-quota-definition-guid"))

				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

				Expect(fakeCloudControllerClient.GetOrganizationCallCount()).To(Equal(1))
				guid := fakeCloudControllerClient.GetOrganizationArgsForCall(0)
				Expect(guid).To(Equal("some-org-guid"))
			})
		})

		When("the org does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationReturns(
					ccv2.Organization{},
					ccv2.Warnings{"warning-1", "warning-2"},
					ccerror.ResourceNotFoundError{},
				)
			})

			It("returns warnings and OrganizationNotFoundError", func() {
				Expect(err).To(MatchError(actionerror.OrganizationNotFoundError{GUID: "some-org-guid"}))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})

		When("client returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some get org error")
				fakeCloudControllerClient.GetOrganizationReturns(
					ccv2.Organization{},
					ccv2.Warnings{"warning-1", "warning-2"},
					expectedErr,
				)
			})

			It("returns warnings and the error", func() {
				Expect(err).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})
	})

	Describe("GetOrganizationByName", func() {
		var (
			org      Organization
			warnings Warnings
			err      error
		)

		JustBeforeEach(func() {
			org, warnings, err = actor.GetOrganizationByName("some-org")
		})

		When("the org exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{
						{GUID: "some-org-guid"},
					},
					ccv2.Warnings{"warning-1", "warning-2"},
					nil)
			})

			It("returns the org and all warnings", func() {
				Expect(org.GUID).To(Equal("some-org-guid"))
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

				Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
				filters := fakeCloudControllerClient.GetOrganizationsArgsForCall(0)
				Expect(filters).To(Equal(
					[]ccv2.Filter{{
						Type:     constant.NameFilter,
						Operator: constant.EqualOperator,
						Values:   []string{"some-org"},
					}}))
			})
		})

		When("the org does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{},
					nil,
					nil,
				)
			})

			It("returns OrganizationNotFoundError", func() {
				Expect(err).To(MatchError(actionerror.OrganizationNotFoundError{Name: "some-org"}))
			})
		})

		When("multiple orgs exist with the same name", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{
						{GUID: "org-1-guid"},
						{GUID: "org-2-guid"},
					},
					nil,
					nil,
				)
			})

			It("returns MultipleOrganizationsFoundError", func() {
				Expect(err).To(MatchError("Organization name 'some-org' matches multiple GUIDs: org-1-guid, org-2-guid"))
			})
		})

		When("an error is encountered", func() {
			var returnedErr error

			BeforeEach(func() {
				returnedErr = errors.New("get-orgs-error")
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{},
					ccv2.Warnings{
						"warning-1",
						"warning-2",
					},
					returnedErr,
				)
			})

			It("returns the error and all warnings", func() {
				Expect(err).To(MatchError(returnedErr))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})
	})

	Describe("GrantOrgManagerByUsername", func() {
		var (
			guid     string
			username string
			warnings Warnings
			err      error
		)
		JustBeforeEach(func() {
			warnings, err = actor.GrantOrgManagerByUsername(guid, username)
		})

		Context("when making the user an org manager succeeds", func() {
			BeforeEach(func() {
				guid = "some-guid"
				username = "some-user"

				fakeCloudControllerClient.UpdateOrganizationManagerByUsernameReturns(
					ccv2.Warnings{"warning-1", "warning-2"},
					nil,
				)
			})

			It("returns warnings", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeCloudControllerClient.UpdateOrganizationManagerByUsernameCallCount()).To(Equal(1))
				orgGuid, user := fakeCloudControllerClient.UpdateOrganizationManagerByUsernameArgsForCall(0)
				Expect(orgGuid).To(Equal("some-guid"))
				Expect(user).To(Equal("some-user"))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})

		Context("when making the user an org manager fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UpdateOrganizationManagerByUsernameReturns(
					ccv2.Warnings{"warning-1", "warning-2"},
					errors.New("some-error"),
				)
			})

			It("returns the error and all warnings", func() {
				Expect(err).To(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				Expect(err).To(MatchError("some-error"))
			})
		})
	})

	Describe("CreateOrganization", func() {
		var (
			org      Organization
			warnings Warnings
			err      error
		)

		JustBeforeEach(func() {
			org, warnings, err = actor.CreateOrganization("some-org", "quota-name")
		})

		Context("the organization is created successfully", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationQuotaByNameReturns(
					ccv2.OrganizationQuota{
						GUID: "some-quota-definition-guid",
						Name: "quota-name",
					},
					ccv2.Warnings{"warning-1", "warning-2"},
					nil)

				fakeCloudControllerClient.CreateOrganizationReturns(
					ccv2.Organization{
						GUID:                "some-org-guid",
						Name:                "some-org",
						QuotaDefinitionGUID: "some-quota-definition-guid",
					},
					ccv2.Warnings{"warning-1", "warning-2"},
					nil)
			})

			It("returns the org and all warnings", func() {
				Expect(err).ToNot(HaveOccurred())

				Expect(org.GUID).To(Equal("some-org-guid"))
				Expect(org.Name).To(Equal("some-org"))
				Expect(org.QuotaDefinitionGUID).To(Equal("some-quota-definition-guid"))

				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

				Expect(fakeCloudControllerClient.GetOrganizationQuotaByNameCallCount()).To(Equal(1))
				quotaName := fakeCloudControllerClient.GetOrganizationQuotaByNameArgsForCall(0)
				Expect(quotaName).To(Equal("quota-name"))

				Expect(fakeCloudControllerClient.CreateOrganizationCallCount()).To(Equal(1))
				orgName, quotaGUID := fakeCloudControllerClient.CreateOrganizationArgsForCall(0)
				Expect(orgName).To(Equal("some-org"))
				Expect(quotaGUID).To(Equal("some-quota-definition-guid"))
			})
		})
	})

	Describe("DeleteOrganization", func() {
		var (
			warnings     Warnings
			deleteOrgErr error
			job          ccv2.Job
		)

		JustBeforeEach(func() {
			warnings, deleteOrgErr = actor.DeleteOrganization("some-org")
		})

		Context("the organization is deleted successfully", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns([]ccv2.Organization{
					{GUID: "some-org-guid"},
				}, ccv2.Warnings{"get-org-warning"}, nil)

				job = ccv2.Job{
					GUID:   "some-job-guid",
					Status: constant.JobStatusFinished,
				}

				fakeCloudControllerClient.DeleteOrganizationJobReturns(
					job, ccv2.Warnings{"delete-org-warning"}, nil)

				fakeCloudControllerClient.PollJobReturns(ccv2.Warnings{"polling-warnings"}, nil)
			})

			It("returns warnings and deletes the org", func() {
				Expect(warnings).To(ConsistOf("get-org-warning", "delete-org-warning", "polling-warnings"))
				Expect(deleteOrgErr).ToNot(HaveOccurred())

				Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
				filters := fakeCloudControllerClient.GetOrganizationsArgsForCall(0)
				Expect(filters).To(Equal(
					[]ccv2.Filter{{
						Type:     constant.NameFilter,
						Operator: constant.EqualOperator,
						Values:   []string{"some-org"},
					}}))

				Expect(fakeCloudControllerClient.DeleteOrganizationJobCallCount()).To(Equal(1))
				orgGuid := fakeCloudControllerClient.DeleteOrganizationJobArgsForCall(0)
				Expect(orgGuid).To(Equal("some-org-guid"))

				Expect(fakeCloudControllerClient.PollJobCallCount()).To(Equal(1))
				job := fakeCloudControllerClient.PollJobArgsForCall(0)
				Expect(job.GUID).To(Equal("some-job-guid"))
			})
		})

		When("getting the org returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{},
					ccv2.Warnings{
						"get-org-warning",
					},
					nil,
				)
			})

			It("returns an error and all warnings", func() {
				Expect(warnings).To(ConsistOf("get-org-warning"))
				Expect(deleteOrgErr).To(MatchError(actionerror.OrganizationNotFoundError{
					Name: "some-org",
				}))
			})
		})

		When("the delete returns an error", func() {
			var returnedErr error

			BeforeEach(func() {
				returnedErr = errors.New("delete-org-error")

				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{{GUID: "org-1-guid"}},
					ccv2.Warnings{
						"get-org-warning",
					},
					nil,
				)

				fakeCloudControllerClient.DeleteOrganizationJobReturns(
					ccv2.Job{},
					ccv2.Warnings{"delete-org-warning"},
					returnedErr)
			})

			It("returns the error and all warnings", func() {
				Expect(deleteOrgErr).To(MatchError(returnedErr))
				Expect(warnings).To(ConsistOf("get-org-warning", "delete-org-warning"))
			})
		})

		When("the job polling has an error", func() {
			var expectedErr error
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns([]ccv2.Organization{
					{GUID: "some-org-guid"},
				}, ccv2.Warnings{"get-org-warning"}, nil)

				fakeCloudControllerClient.DeleteOrganizationJobReturns(
					ccv2.Job{}, ccv2.Warnings{"delete-org-warning"}, nil)

				expectedErr = errors.New("Never expected, by anyone")
				fakeCloudControllerClient.PollJobReturns(ccv2.Warnings{"polling-warnings"}, expectedErr)
			})

			It("returns the error from job polling", func() {
				Expect(warnings).To(ConsistOf("get-org-warning", "delete-org-warning", "polling-warnings"))
				Expect(deleteOrgErr).To(MatchError(expectedErr))
			})
		})
	})

	Describe("GetOrganizations", func() {
		var (
			orgs     []Organization
			warnings Warnings
			err      error
		)

		JustBeforeEach(func() {
			orgs, warnings, err = actor.GetOrganizations()
		})

		When("there are multiple organizations", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{
						{
							Name: "some-org-1",
						},
						{
							Name: "some-org-2",
						},
					},
					ccv2.Warnings{"warning-1", "warning-2"},
					nil)
			})

			It("returns the org and all warnings", func() {
				Expect(err).ToNot(HaveOccurred())

				Expect(orgs).To(HaveLen(2))
				Expect(orgs[0].Name).To(Equal("some-org-1"))
				Expect(orgs[1].Name).To(Equal("some-org-2"))

				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

				Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
				queriesArg := fakeCloudControllerClient.GetOrganizationsArgsForCall(0)
				Expect(queriesArg).To(BeNil())
			})
		})

		When("there are no orgs", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{},
					ccv2.Warnings{"warning-1", "warning-2"},
					nil,
				)
			})

			It("returns warnings and an empty list of orgs", func() {
				Expect(err).ToNot(HaveOccurred())

				Expect(orgs).To(HaveLen(0))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})

		When("client returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some get org error")
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv2.Organization{},
					ccv2.Warnings{"warning-1", "warning-2"},
					expectedErr,
				)
			})

			It("returns warnings and the error", func() {
				Expect(err).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})
	})
})
