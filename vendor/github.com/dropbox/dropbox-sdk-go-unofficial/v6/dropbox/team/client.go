// Copyright (c) Dropbox, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package team

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/async"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/auth"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/file_properties"
)

// Client interface describes all routes in this namespace
type Client interface {
	// DevicesListMemberDevices : List all device sessions of a team's member.
	DevicesListMemberDevices(arg *ListMemberDevicesArg) (res *ListMemberDevicesResult, err error)
	// DevicesListMembersDevices : List all device sessions of a team.
	// Permission : Team member file access.
	DevicesListMembersDevices(arg *ListMembersDevicesArg) (res *ListMembersDevicesResult, err error)
	// DevicesListTeamDevices : List all device sessions of a team. Permission :
	// Team member file access.
	// Deprecated:
	DevicesListTeamDevices(arg *ListTeamDevicesArg) (res *ListTeamDevicesResult, err error)
	// DevicesRevokeDeviceSession : Revoke a device session of a team's member.
	DevicesRevokeDeviceSession(arg *RevokeDeviceSessionArg) (err error)
	// DevicesRevokeDeviceSessionBatch : Revoke a list of device sessions of
	// team members.
	DevicesRevokeDeviceSessionBatch(arg *RevokeDeviceSessionBatchArg) (res *RevokeDeviceSessionBatchResult, err error)
	// FeaturesGetValues : Get the values for one or more features. This route
	// allows you to check your account's capability for what feature you can
	// access or what value you have for certain features. Permission : Team
	// information.
	FeaturesGetValues(arg *FeaturesGetValuesBatchArg) (res *FeaturesGetValuesBatchResult, err error)
	// GetInfo : Retrieves information about a team.
	GetInfo() (res *TeamGetInfoResult, err error)
	// GroupsCreate : Creates a new, empty group, with a requested name.
	// Permission : Team member management.
	GroupsCreate(arg *GroupCreateArg) (res *GroupFullInfo, err error)
	// GroupsDelete : Deletes a group. The group is deleted immediately. However
	// the revoking of group-owned resources may take additional time. Use the
	// `groupsJobStatusGet` to determine whether this process has completed.
	// Permission : Team member management.
	GroupsDelete(arg *GroupSelector) (res *async.LaunchEmptyResult, err error)
	// GroupsGetInfo : Retrieves information about one or more groups. Note that
	// the optional field `GroupFullInfo.members` is not returned for
	// system-managed groups. Permission : Team Information.
	GroupsGetInfo(arg *GroupsSelector) (res []*GroupsGetInfoItem, err error)
	// GroupsJobStatusGet : Once an async_job_id is returned from
	// `groupsDelete`, `groupsMembersAdd` , or `groupsMembersRemove` use this
	// method to poll the status of granting/revoking group members' access to
	// group-owned resources. Permission : Team member management.
	GroupsJobStatusGet(arg *async.PollArg) (res *async.PollEmptyResult, err error)
	// GroupsList : Lists groups on a team. Permission : Team Information.
	GroupsList(arg *GroupsListArg) (res *GroupsListResult, err error)
	// GroupsListContinue : Once a cursor has been retrieved from `groupsList`,
	// use this to paginate through all groups. Permission : Team Information.
	GroupsListContinue(arg *GroupsListContinueArg) (res *GroupsListResult, err error)
	// GroupsMembersAdd : Adds members to a group. The members are added
	// immediately. However the granting of group-owned resources may take
	// additional time. Use the `groupsJobStatusGet` to determine whether this
	// process has completed. Permission : Team member management.
	GroupsMembersAdd(arg *GroupMembersAddArg) (res *GroupMembersChangeResult, err error)
	// GroupsMembersList : Lists members of a group. Permission : Team
	// Information.
	GroupsMembersList(arg *GroupsMembersListArg) (res *GroupsMembersListResult, err error)
	// GroupsMembersListContinue : Once a cursor has been retrieved from
	// `groupsMembersList`, use this to paginate through all members of the
	// group. Permission : Team information.
	GroupsMembersListContinue(arg *GroupsMembersListContinueArg) (res *GroupsMembersListResult, err error)
	// GroupsMembersRemove : Removes members from a group. The members are
	// removed immediately. However the revoking of group-owned resources may
	// take additional time. Use the `groupsJobStatusGet` to determine whether
	// this process has completed. This method permits removing the only owner
	// of a group, even in cases where this is not possible via the web client.
	// Permission : Team member management.
	GroupsMembersRemove(arg *GroupMembersRemoveArg) (res *GroupMembersChangeResult, err error)
	// GroupsMembersSetAccessType : Sets a member's access type in a group.
	// Permission : Team member management.
	GroupsMembersSetAccessType(arg *GroupMembersSetAccessTypeArg) (res []*GroupsGetInfoItem, err error)
	// GroupsUpdate : Updates a group's name and/or external ID. Permission :
	// Team member management.
	GroupsUpdate(arg *GroupUpdateArgs) (res *GroupFullInfo, err error)
	// LegalHoldsCreatePolicy : Creates new legal hold policy. Note: Legal Holds
	// is a paid add-on. Not all teams have the feature. Permission : Team
	// member file access.
	LegalHoldsCreatePolicy(arg *LegalHoldsPolicyCreateArg) (res *LegalHoldPolicy, err error)
	// LegalHoldsGetPolicy : Gets a legal hold by Id. Note: Legal Holds is a
	// paid add-on. Not all teams have the feature. Permission : Team member
	// file access.
	LegalHoldsGetPolicy(arg *LegalHoldsGetPolicyArg) (res *LegalHoldPolicy, err error)
	// LegalHoldsListHeldRevisions : List the file metadata that's under the
	// hold. Note: Legal Holds is a paid add-on. Not all teams have the feature.
	// Permission : Team member file access.
	LegalHoldsListHeldRevisions(arg *LegalHoldsListHeldRevisionsArg) (res *LegalHoldsListHeldRevisionResult, err error)
	// LegalHoldsListHeldRevisionsContinue : Continue listing the file metadata
	// that's under the hold. Note: Legal Holds is a paid add-on. Not all teams
	// have the feature. Permission : Team member file access.
	LegalHoldsListHeldRevisionsContinue(arg *LegalHoldsListHeldRevisionsContinueArg) (res *LegalHoldsListHeldRevisionResult, err error)
	// LegalHoldsListPolicies : Lists legal holds on a team. Note: Legal Holds
	// is a paid add-on. Not all teams have the feature. Permission : Team
	// member file access.
	LegalHoldsListPolicies(arg *LegalHoldsListPoliciesArg) (res *LegalHoldsListPoliciesResult, err error)
	// LegalHoldsReleasePolicy : Releases a legal hold by Id. Note: Legal Holds
	// is a paid add-on. Not all teams have the feature. Permission : Team
	// member file access.
	LegalHoldsReleasePolicy(arg *LegalHoldsPolicyReleaseArg) (err error)
	// LegalHoldsUpdatePolicy : Updates a legal hold. Note: Legal Holds is a
	// paid add-on. Not all teams have the feature. Permission : Team member
	// file access.
	LegalHoldsUpdatePolicy(arg *LegalHoldsPolicyUpdateArg) (res *LegalHoldPolicy, err error)
	// LinkedAppsListMemberLinkedApps : List all linked applications of the team
	// member. Note, this endpoint does not list any team-linked applications.
	LinkedAppsListMemberLinkedApps(arg *ListMemberAppsArg) (res *ListMemberAppsResult, err error)
	// LinkedAppsListMembersLinkedApps : List all applications linked to the
	// team members' accounts. Note, this endpoint does not list any team-linked
	// applications.
	LinkedAppsListMembersLinkedApps(arg *ListMembersAppsArg) (res *ListMembersAppsResult, err error)
	// LinkedAppsListTeamLinkedApps : List all applications linked to the team
	// members' accounts. Note, this endpoint doesn't list any team-linked
	// applications.
	// Deprecated:
	LinkedAppsListTeamLinkedApps(arg *ListTeamAppsArg) (res *ListTeamAppsResult, err error)
	// LinkedAppsRevokeLinkedApp : Revoke a linked application of the team
	// member.
	LinkedAppsRevokeLinkedApp(arg *RevokeLinkedApiAppArg) (err error)
	// LinkedAppsRevokeLinkedAppBatch : Revoke a list of linked applications of
	// the team members.
	LinkedAppsRevokeLinkedAppBatch(arg *RevokeLinkedApiAppBatchArg) (res *RevokeLinkedAppBatchResult, err error)
	// MemberSpaceLimitsExcludedUsersAdd : Add users to member space limits
	// excluded users list.
	MemberSpaceLimitsExcludedUsersAdd(arg *ExcludedUsersUpdateArg) (res *ExcludedUsersUpdateResult, err error)
	// MemberSpaceLimitsExcludedUsersList : List member space limits excluded
	// users.
	MemberSpaceLimitsExcludedUsersList(arg *ExcludedUsersListArg) (res *ExcludedUsersListResult, err error)
	// MemberSpaceLimitsExcludedUsersListContinue : Continue listing member
	// space limits excluded users.
	MemberSpaceLimitsExcludedUsersListContinue(arg *ExcludedUsersListContinueArg) (res *ExcludedUsersListResult, err error)
	// MemberSpaceLimitsExcludedUsersRemove : Remove users from member space
	// limits excluded users list.
	MemberSpaceLimitsExcludedUsersRemove(arg *ExcludedUsersUpdateArg) (res *ExcludedUsersUpdateResult, err error)
	// MemberSpaceLimitsGetCustomQuota : Get users custom quota. A maximum of
	// 1000 members can be specified in a single call. Note: to apply a custom
	// space limit, a team admin needs to set a member space limit for the team
	// first. (the team admin can check the settings here:
	// https://www.dropbox.com/team/admin/settings/space).
	MemberSpaceLimitsGetCustomQuota(arg *CustomQuotaUsersArg) (res []*CustomQuotaResult, err error)
	// MemberSpaceLimitsRemoveCustomQuota : Remove users custom quota. A maximum
	// of 1000 members can be specified in a single call. Note: to apply a
	// custom space limit, a team admin needs to set a member space limit for
	// the team first. (the team admin can check the settings here:
	// https://www.dropbox.com/team/admin/settings/space).
	MemberSpaceLimitsRemoveCustomQuota(arg *CustomQuotaUsersArg) (res []*RemoveCustomQuotaResult, err error)
	// MemberSpaceLimitsSetCustomQuota : Set users custom quota. Custom quota
	// has to be at least 2GB. A maximum of 1000 members can be specified in a
	// single call. Note: to apply a custom space limit, a team admin needs to
	// set a member space limit for the team first. (the team admin can check
	// the settings here: https://www.dropbox.com/team/admin/settings/space).
	MemberSpaceLimitsSetCustomQuota(arg *SetCustomQuotaArg) (res []*CustomQuotaResult, err error)
	// MembersAdd : Adds members to a team. Permission : Team member management
	// A maximum of 20 members can be specified in a single call. If no Dropbox
	// account exists with the email address specified, a new Dropbox account
	// will be created with the given email address, and that account will be
	// invited to the team. If a personal Dropbox account exists with the email
	// address specified in the call, this call will create a placeholder
	// Dropbox account for the user on the team and send an email inviting the
	// user to migrate their existing personal account onto the team. Team
	// member management apps are required to set an initial given_name and
	// surname for a user to use in the team invitation and for 'Perform as team
	// member' actions taken on the user before they become 'active'.
	MembersAdd(arg *MembersAddArg) (res *MembersAddLaunch, err error)
	// MembersAdd : Adds members to a team. Permission : Team member management
	// A maximum of 20 members can be specified in a single call. If no Dropbox
	// account exists with the email address specified, a new Dropbox account
	// will be created with the given email address, and that account will be
	// invited to the team. If a personal Dropbox account exists with the email
	// address specified in the call, this call will create a placeholder
	// Dropbox account for the user on the team and send an email inviting the
	// user to migrate their existing personal account onto the team.
	MembersAddV2(arg *MembersAddV2Arg) (res *MembersAddLaunchV2Result, err error)
	// MembersAddJobStatusGet : Once an async_job_id is returned from
	// `membersAdd` , use this to poll the status of the asynchronous request.
	// Permission : Team member management.
	MembersAddJobStatusGet(arg *async.PollArg) (res *MembersAddJobStatus, err error)
	// MembersAddJobStatusGet : Once an async_job_id is returned from
	// `membersAdd` , use this to poll the status of the asynchronous request.
	// Permission : Team member management.
	MembersAddJobStatusGetV2(arg *async.PollArg) (res *MembersAddJobStatusV2Result, err error)
	// MembersBulkSuspend : Launch a bulk suspend job. The server enforces a
	// maximum of 500 members.
	MembersBulkSuspend(arg *BulkSuspendArg) (res *async.LaunchResultBase, err error)
	// MembersBulkSuspendJobStatusCheck : Poll a previously launched bulk
	// suspend job.
	MembersBulkSuspendJobStatusCheck(arg *async.PollArg) (res *BulkSuspendJobStatus, err error)
	// MembersDeleteFormerMemberFiles : Permanently delete the files of a user
	// who has been removed from the team. After permanent deletion, those files
	// will not be available to be transferred to another team member.
	// Permission : Team member management Exactly one of team_member_id, email,
	// or external_id must be provided to identify the user account.
	MembersDeleteFormerMemberFiles(arg *MembersFormerMemberArg) (err error)
	// MembersDeleteProfilePhoto : Deletes a team member's profile photo.
	// Permission : Team member management.
	MembersDeleteProfilePhoto(arg *MembersDeleteProfilePhotoArg) (res *TeamMemberInfo, err error)
	// MembersDeleteProfilePhoto : Deletes a team member's profile photo.
	// Permission : Team member management.
	MembersDeleteProfilePhotoV2(arg *MembersDeleteProfilePhotoArg) (res *TeamMemberInfoV2Result, err error)
	// MembersGetAvailableTeamMemberRoles : Get available TeamMemberRoles for
	// the connected team. To be used with `membersSetAdminPermissions`.
	// Permission : Team member management.
	MembersGetAvailableTeamMemberRoles() (res *MembersGetAvailableTeamMemberRolesResult, err error)
	// MembersGetInfo : Returns information about multiple team members.
	// Permission : Team information This endpoint will return
	// `MembersGetInfoItem.id_not_found`, for IDs (or emails) that cannot be
	// matched to a valid team member.
	MembersGetInfo(arg *MembersGetInfoArgs) (res []*MembersGetInfoItem, err error)
	// MembersGetInfo : Returns information about multiple team members.
	// Permission : Team information This endpoint will return
	// `MembersGetInfoItem.id_not_found`, for IDs (or emails) that cannot be
	// matched to a valid team member.
	MembersGetInfoV2(arg *MembersGetInfoV2Arg) (res *MembersGetInfoV2Result, err error)
	// MembersList : Lists members of a team. Permission : Team information.
	MembersList(arg *MembersListArg) (res *MembersListResult, err error)
	// MembersList : Lists members of a team. Permission : Team information.
	MembersListV2(arg *MembersListArg) (res *MembersListV2Result, err error)
	// MembersListContinue : Once a cursor has been retrieved from
	// `membersList`, use this to paginate through all team members. Permission
	// : Team information.
	MembersListContinue(arg *MembersListContinueArg) (res *MembersListResult, err error)
	// MembersListContinue : Once a cursor has been retrieved from
	// `membersList`, use this to paginate through all team members. Permission
	// : Team information.
	MembersListContinueV2(arg *MembersListContinueArg) (res *MembersListV2Result, err error)
	// MembersMoveFormerMemberFiles : Moves removed member's files to a
	// different member. This endpoint initiates an asynchronous job. To obtain
	// the final result of the job, the client should periodically poll
	// `membersMoveFormerMemberFilesJobStatusCheck`. Permission : Team member
	// management.
	MembersMoveFormerMemberFiles(arg *MembersDataTransferArg) (res *async.LaunchEmptyResult, err error)
	// MembersMoveFormerMemberFilesJobStatusCheck : Once an async_job_id is
	// returned from `membersMoveFormerMemberFiles` , use this to poll the
	// status of the asynchronous request. Permission : Team member management.
	MembersMoveFormerMemberFilesJobStatusCheck(arg *async.PollArg) (res *async.PollEmptyResult, err error)
	// MembersRecover : Recover a deleted member. Permission : Team member
	// management Exactly one of team_member_id, email, or external_id must be
	// provided to identify the user account.
	MembersRecover(arg *MembersRecoverArg) (err error)
	// MembersRemove : Removes a member from a team. Permission : Team member
	// management Exactly one of team_member_id, email, or external_id must be
	// provided to identify the user account. Accounts can be recovered via
	// `membersRecover` for a 7 day period or until the account has been
	// permanently deleted or transferred to another account (whichever comes
	// first). Calling `membersAdd` while a user is still recoverable on your
	// team will return with `MemberAddResult.user_already_on_team`. Accounts
	// can have their files transferred via the admin console for a limited
	// time, based on the version history length associated with the team (180
	// days for most teams). Accounts can have their stacks transferred through
	// the admin console. This only transfers stacks that they have created.
	// This endpoint may initiate an asynchronous job. To obtain the final
	// result of the job, the client should periodically poll
	// `membersRemoveJobStatusGet`.
	MembersRemove(arg *MembersRemoveArg) (res *async.LaunchEmptyResult, err error)
	// MembersRemoveJobStatusGet : Once an async_job_id is returned from
	// `membersRemove` , use this to poll the status of the asynchronous
	// request. Permission : Team member management.
	MembersRemoveJobStatusGet(arg *async.PollArg) (res *async.PollEmptyResult, err error)
	// MembersSecondaryEmailsAdd : Add secondary emails to users. Permission :
	// Team member management. Emails that are on verified domains will be
	// verified automatically. For each email address not on a verified domain a
	// verification email will be sent.
	MembersSecondaryEmailsAdd(arg *AddSecondaryEmailsArg) (res *AddSecondaryEmailsResult, err error)
	// MembersSecondaryEmailsDelete : Delete secondary emails from users
	// Permission : Team member management. Users will be notified of deletions
	// of verified secondary emails at both the secondary email and their
	// primary email.
	MembersSecondaryEmailsDelete(arg *DeleteSecondaryEmailsArg) (res *DeleteSecondaryEmailsResult, err error)
	// MembersSecondaryEmailsResendVerificationEmails : Resend secondary email
	// verification emails. Permission : Team member management.
	MembersSecondaryEmailsResendVerificationEmails(arg *ResendVerificationEmailArg) (res *ResendVerificationEmailResult, err error)
	// MembersSendWelcomeEmail : Sends welcome email to pending team member.
	// Permission : Team member management Exactly one of team_member_id, email,
	// or external_id must be provided to identify the user account. No-op if
	// team member is not pending.
	MembersSendWelcomeEmail(arg *UserSelectorArg) (err error)
	// MembersSetAdminPermissions : Updates a team member's permissions.
	// Permission : Team member management.
	MembersSetAdminPermissions(arg *MembersSetPermissionsArg) (res *MembersSetPermissionsResult, err error)
	// MembersSetAdminPermissions : Updates a team member's permissions.
	// Permission : Team member management.
	MembersSetAdminPermissionsV2(arg *MembersSetPermissions2Arg) (res *MembersSetPermissions2Result, err error)
	// MembersSetProfile : Updates a team member's profile. Permission : Team
	// member management.
	MembersSetProfile(arg *MembersSetProfileArg) (res *TeamMemberInfo, err error)
	// MembersSetProfile : Updates a team member's profile. Permission : Team
	// member management.
	MembersSetProfileV2(arg *MembersSetProfileArg) (res *TeamMemberInfoV2Result, err error)
	// MembersSetProfilePhoto : Updates a team member's profile photo.
	// Permission : Team member management.
	MembersSetProfilePhoto(arg *MembersSetProfilePhotoArg) (res *TeamMemberInfo, err error)
	// MembersSetProfilePhoto : Updates a team member's profile photo.
	// Permission : Team member management.
	MembersSetProfilePhotoV2(arg *MembersSetProfilePhotoArg) (res *TeamMemberInfoV2Result, err error)
	// MembersSuspend : Suspend a member from a team. Permission : Team member
	// management Exactly one of team_member_id, email, or external_id must be
	// provided to identify the user account.
	MembersSuspend(arg *MembersDeactivateArg) (err error)
	// MembersUnsuspend : Unsuspend a member from a team. Permission : Team
	// member management Exactly one of team_member_id, email, or external_id
	// must be provided to identify the user account.
	MembersUnsuspend(arg *MembersUnsuspendArg) (err error)
	// NamespacesList : Returns a list of all team-accessible namespaces. This
	// list includes team folders, shared folders containing team members, team
	// members' home namespaces, and team members' app folders. Home namespaces
	// and app folders are always owned by this team or members of the team, but
	// shared folders may be owned by other users or other teams. Duplicates may
	// occur in the list.
	NamespacesList(arg *TeamNamespacesListArg) (res *TeamNamespacesListResult, err error)
	// NamespacesListContinue : Once a cursor has been retrieved from
	// `namespacesList`, use this to paginate through all team-accessible
	// namespaces. Duplicates may occur in the list.
	NamespacesListContinue(arg *TeamNamespacesListContinueArg) (res *TeamNamespacesListResult, err error)
	// PropertiesTemplateAdd : Permission : Team member file access.
	// Deprecated:
	PropertiesTemplateAdd(arg *file_properties.AddTemplateArg) (res *file_properties.AddTemplateResult, err error)
	// PropertiesTemplateGet : Permission : Team member file access. The scope
	// for the route is files.team_metadata.write.
	// Deprecated:
	PropertiesTemplateGet(arg *file_properties.GetTemplateArg) (res *file_properties.GetTemplateResult, err error)
	// ReportsGetActivity : Retrieves reporting data about a team's user
	// activity. Deprecated: Will be removed on July 1st 2021.
	// Deprecated:
	ReportsGetActivity(arg *DateRange) (res *GetActivityReport, err error)
	// ReportsGetDevices : Retrieves reporting data about a team's linked
	// devices. Deprecated: Will be removed on July 1st 2021.
	// Deprecated:
	ReportsGetDevices(arg *DateRange) (res *GetDevicesReport, err error)
	// ReportsGetMembership : Retrieves reporting data about a team's
	// membership. Deprecated: Will be removed on July 1st 2021.
	// Deprecated:
	ReportsGetMembership(arg *DateRange) (res *GetMembershipReport, err error)
	// ReportsGetStorage : Retrieves reporting data about a team's storage
	// usage. Deprecated: Will be removed on July 1st 2021.
	// Deprecated:
	ReportsGetStorage(arg *DateRange) (res *GetStorageReport, err error)
	// SharingAllowlistAdd : Endpoint adds Approve List entries. Changes are
	// effective immediately. Changes are committed in transaction. In case of
	// single validation error - all entries are rejected. Valid domains
	// (RFC-1034/5) and emails (RFC-5322/822) are accepted. Added entries cannot
	// overflow limit of 10000 entries per team. Maximum 100 entries per call is
	// allowed.
	SharingAllowlistAdd(arg *SharingAllowlistAddArgs) (res *SharingAllowlistAddResponse, err error)
	// SharingAllowlistList : Lists Approve List entries for given team, from
	// newest to oldest, returning up to `limit` entries at a time. If there are
	// more than `limit` entries associated with the current team, more can be
	// fetched by passing the returned `cursor` to
	// `sharingAllowlistListContinue`.
	SharingAllowlistList(arg *SharingAllowlistListArg) (res *SharingAllowlistListResponse, err error)
	// SharingAllowlistListContinue : Lists entries associated with given team,
	// starting from a the cursor. See `sharingAllowlistList`.
	SharingAllowlistListContinue(arg *SharingAllowlistListContinueArg) (res *SharingAllowlistListResponse, err error)
	// SharingAllowlistRemove : Endpoint removes Approve List entries. Changes
	// are effective immediately. Changes are committed in transaction. In case
	// of single validation error - all entries are rejected. Valid domains
	// (RFC-1034/5) and emails (RFC-5322/822) are accepted. Entries being
	// removed have to be present on the list. Maximum 1000 entries per call is
	// allowed.
	SharingAllowlistRemove(arg *SharingAllowlistRemoveArgs) (res *SharingAllowlistRemoveResponse, err error)
	// TeamFolderActivate : Sets an archived team folder's status to active.
	// Permission : Team member file access.
	TeamFolderActivate(arg *TeamFolderIdArg) (res *TeamFolderMetadata, err error)
	// TeamFolderArchive : Sets an active team folder's status to archived and
	// removes all folder and file members. This endpoint cannot be used for
	// teams that have a shared team space. This route will either finish
	// synchronously, or return a job ID and do the async archive job in
	// background. Please use team_folder/archive/check to check the job status.
	// Permission : Team member file access.
	TeamFolderArchive(arg *TeamFolderArchiveArg) (res *TeamFolderArchiveLaunch, err error)
	// TeamFolderArchiveCheck : Returns the status of an asynchronous job for
	// archiving a team folder. The job may show '.tag' as complete, but the
	// team folder could still be in the process of archiving (indicated by
	// `TeamFolderMetadata.status` with 'archive_in_progress'). To confirm that
	// the team folder is fully archived, check the field
	// `TeamFolderMetadata.status` in the response for the value 'archived'.
	// Permission : Team member file access.
	TeamFolderArchiveCheck(arg *async.PollArg) (res *TeamFolderArchiveJobStatus, err error)
	// TeamFolderCreate : Creates a new, active, team folder with no members.
	// This endpoint can only be used for teams that do not already have a
	// shared team space. Permission : Team member file access.
	TeamFolderCreate(arg *TeamFolderCreateArg) (res *TeamFolderMetadata, err error)
	// TeamFolderGetInfo : Retrieves metadata for team folders. Permission :
	// Team member file access.
	TeamFolderGetInfo(arg *TeamFolderIdListArg) (res []*TeamFolderGetInfoItem, err error)
	// TeamFolderList : Lists all team folders. Permission : Team member file
	// access.
	TeamFolderList(arg *TeamFolderListArg) (res *TeamFolderListResult, err error)
	// TeamFolderListContinue : Once a cursor has been retrieved from
	// `teamFolderList`, use this to paginate through all team folders.
	// Permission : Team member file access.
	TeamFolderListContinue(arg *TeamFolderListContinueArg) (res *TeamFolderListResult, err error)
	// TeamFolderPermanentlyDelete : Permanently deletes an archived team
	// folder. This endpoint cannot be used for teams that have a shared team
	// space. Permission : Team member file access.
	TeamFolderPermanentlyDelete(arg *TeamFolderIdArg) (err error)
	// TeamFolderRename : Changes an active team folder's name. Permission :
	// Team member file access.
	TeamFolderRename(arg *TeamFolderRenameArg) (res *TeamFolderMetadata, err error)
	// TeamFolderRestore : Sets an inactive team folder's status to active.
	// Permission: Team member file access.
	TeamFolderRestore(arg *TeamFolderIdArg) (res *TeamFolderMetadata, err error)
	// TeamFolderUpdateSyncSettings : Updates the sync settings on a team folder
	// or its contents.  Use of this endpoint requires that the team has team
	// selective sync enabled.
	TeamFolderUpdateSyncSettings(arg *TeamFolderUpdateSyncSettingsArg) (res *TeamFolderMetadata, err error)
	// TokenGetAuthenticatedAdmin : Returns the member profile of the admin who
	// generated the team access token used to make the call.
	TokenGetAuthenticatedAdmin() (res *TokenGetAuthenticatedAdminResult, err error)
}

// ContextClient interface describes all routes in this namespace with context support
type ContextClient interface {
	Client
	// DevicesListMemberDevicesContext : List all device sessions of a team's
	// member.
	DevicesListMemberDevicesContext(ctx context.Context, arg *ListMemberDevicesArg) (res *ListMemberDevicesResult, err error)
	// DevicesListMembersDevicesContext : List all device sessions of a team.
	// Permission : Team member file access.
	DevicesListMembersDevicesContext(ctx context.Context, arg *ListMembersDevicesArg) (res *ListMembersDevicesResult, err error)
	// DevicesListTeamDevicesContext : List all device sessions of a team.
	// Permission : Team member file access.
	// Deprecated:
	DevicesListTeamDevicesContext(ctx context.Context, arg *ListTeamDevicesArg) (res *ListTeamDevicesResult, err error)
	// DevicesRevokeDeviceSessionContext : Revoke a device session of a team's
	// member.
	DevicesRevokeDeviceSessionContext(ctx context.Context, arg *RevokeDeviceSessionArg) (err error)
	// DevicesRevokeDeviceSessionBatchContext : Revoke a list of device sessions
	// of team members.
	DevicesRevokeDeviceSessionBatchContext(ctx context.Context, arg *RevokeDeviceSessionBatchArg) (res *RevokeDeviceSessionBatchResult, err error)
	// FeaturesGetValuesContext : Get the values for one or more features. This
	// route allows you to check your account's capability for what feature you
	// can access or what value you have for certain features. Permission : Team
	// information.
	FeaturesGetValuesContext(ctx context.Context, arg *FeaturesGetValuesBatchArg) (res *FeaturesGetValuesBatchResult, err error)
	// GetInfoContext : Retrieves information about a team.
	GetInfoContext(ctx context.Context) (res *TeamGetInfoResult, err error)
	// GroupsCreateContext : Creates a new, empty group, with a requested name.
	// Permission : Team member management.
	GroupsCreateContext(ctx context.Context, arg *GroupCreateArg) (res *GroupFullInfo, err error)
	// GroupsDeleteContext : Deletes a group. The group is deleted immediately.
	// However the revoking of group-owned resources may take additional time.
	// Use the `groupsJobStatusGet` to determine whether this process has
	// completed. Permission : Team member management.
	GroupsDeleteContext(ctx context.Context, arg *GroupSelector) (res *async.LaunchEmptyResult, err error)
	// GroupsGetInfoContext : Retrieves information about one or more groups.
	// Note that the optional field `GroupFullInfo.members` is not returned for
	// system-managed groups. Permission : Team Information.
	GroupsGetInfoContext(ctx context.Context, arg *GroupsSelector) (res []*GroupsGetInfoItem, err error)
	// GroupsJobStatusGetContext : Once an async_job_id is returned from
	// `groupsDelete`, `groupsMembersAdd` , or `groupsMembersRemove` use this
	// method to poll the status of granting/revoking group members' access to
	// group-owned resources. Permission : Team member management.
	GroupsJobStatusGetContext(ctx context.Context, arg *async.PollArg) (res *async.PollEmptyResult, err error)
	// GroupsListContext : Lists groups on a team. Permission : Team
	// Information.
	GroupsListContext(ctx context.Context, arg *GroupsListArg) (res *GroupsListResult, err error)
	// GroupsListContinueContext : Once a cursor has been retrieved from
	// `groupsList`, use this to paginate through all groups. Permission : Team
	// Information.
	GroupsListContinueContext(ctx context.Context, arg *GroupsListContinueArg) (res *GroupsListResult, err error)
	// GroupsMembersAddContext : Adds members to a group. The members are added
	// immediately. However the granting of group-owned resources may take
	// additional time. Use the `groupsJobStatusGet` to determine whether this
	// process has completed. Permission : Team member management.
	GroupsMembersAddContext(ctx context.Context, arg *GroupMembersAddArg) (res *GroupMembersChangeResult, err error)
	// GroupsMembersListContext : Lists members of a group. Permission : Team
	// Information.
	GroupsMembersListContext(ctx context.Context, arg *GroupsMembersListArg) (res *GroupsMembersListResult, err error)
	// GroupsMembersListContinueContext : Once a cursor has been retrieved from
	// `groupsMembersList`, use this to paginate through all members of the
	// group. Permission : Team information.
	GroupsMembersListContinueContext(ctx context.Context, arg *GroupsMembersListContinueArg) (res *GroupsMembersListResult, err error)
	// GroupsMembersRemoveContext : Removes members from a group. The members
	// are removed immediately. However the revoking of group-owned resources
	// may take additional time. Use the `groupsJobStatusGet` to determine
	// whether this process has completed. This method permits removing the only
	// owner of a group, even in cases where this is not possible via the web
	// client. Permission : Team member management.
	GroupsMembersRemoveContext(ctx context.Context, arg *GroupMembersRemoveArg) (res *GroupMembersChangeResult, err error)
	// GroupsMembersSetAccessTypeContext : Sets a member's access type in a
	// group. Permission : Team member management.
	GroupsMembersSetAccessTypeContext(ctx context.Context, arg *GroupMembersSetAccessTypeArg) (res []*GroupsGetInfoItem, err error)
	// GroupsUpdateContext : Updates a group's name and/or external ID.
	// Permission : Team member management.
	GroupsUpdateContext(ctx context.Context, arg *GroupUpdateArgs) (res *GroupFullInfo, err error)
	// LegalHoldsCreatePolicyContext : Creates new legal hold policy. Note:
	// Legal Holds is a paid add-on. Not all teams have the feature. Permission
	// : Team member file access.
	LegalHoldsCreatePolicyContext(ctx context.Context, arg *LegalHoldsPolicyCreateArg) (res *LegalHoldPolicy, err error)
	// LegalHoldsGetPolicyContext : Gets a legal hold by Id. Note: Legal Holds
	// is a paid add-on. Not all teams have the feature. Permission : Team
	// member file access.
	LegalHoldsGetPolicyContext(ctx context.Context, arg *LegalHoldsGetPolicyArg) (res *LegalHoldPolicy, err error)
	// LegalHoldsListHeldRevisionsContext : List the file metadata that's under
	// the hold. Note: Legal Holds is a paid add-on. Not all teams have the
	// feature. Permission : Team member file access.
	LegalHoldsListHeldRevisionsContext(ctx context.Context, arg *LegalHoldsListHeldRevisionsArg) (res *LegalHoldsListHeldRevisionResult, err error)
	// LegalHoldsListHeldRevisionsContinueContext : Continue listing the file
	// metadata that's under the hold. Note: Legal Holds is a paid add-on. Not
	// all teams have the feature. Permission : Team member file access.
	LegalHoldsListHeldRevisionsContinueContext(ctx context.Context, arg *LegalHoldsListHeldRevisionsContinueArg) (res *LegalHoldsListHeldRevisionResult, err error)
	// LegalHoldsListPoliciesContext : Lists legal holds on a team. Note: Legal
	// Holds is a paid add-on. Not all teams have the feature. Permission : Team
	// member file access.
	LegalHoldsListPoliciesContext(ctx context.Context, arg *LegalHoldsListPoliciesArg) (res *LegalHoldsListPoliciesResult, err error)
	// LegalHoldsReleasePolicyContext : Releases a legal hold by Id. Note: Legal
	// Holds is a paid add-on. Not all teams have the feature. Permission : Team
	// member file access.
	LegalHoldsReleasePolicyContext(ctx context.Context, arg *LegalHoldsPolicyReleaseArg) (err error)
	// LegalHoldsUpdatePolicyContext : Updates a legal hold. Note: Legal Holds
	// is a paid add-on. Not all teams have the feature. Permission : Team
	// member file access.
	LegalHoldsUpdatePolicyContext(ctx context.Context, arg *LegalHoldsPolicyUpdateArg) (res *LegalHoldPolicy, err error)
	// LinkedAppsListMemberLinkedAppsContext : List all linked applications of
	// the team member. Note, this endpoint does not list any team-linked
	// applications.
	LinkedAppsListMemberLinkedAppsContext(ctx context.Context, arg *ListMemberAppsArg) (res *ListMemberAppsResult, err error)
	// LinkedAppsListMembersLinkedAppsContext : List all applications linked to
	// the team members' accounts. Note, this endpoint does not list any
	// team-linked applications.
	LinkedAppsListMembersLinkedAppsContext(ctx context.Context, arg *ListMembersAppsArg) (res *ListMembersAppsResult, err error)
	// LinkedAppsListTeamLinkedAppsContext : List all applications linked to the
	// team members' accounts. Note, this endpoint doesn't list any team-linked
	// applications.
	// Deprecated:
	LinkedAppsListTeamLinkedAppsContext(ctx context.Context, arg *ListTeamAppsArg) (res *ListTeamAppsResult, err error)
	// LinkedAppsRevokeLinkedAppContext : Revoke a linked application of the
	// team member.
	LinkedAppsRevokeLinkedAppContext(ctx context.Context, arg *RevokeLinkedApiAppArg) (err error)
	// LinkedAppsRevokeLinkedAppBatchContext : Revoke a list of linked
	// applications of the team members.
	LinkedAppsRevokeLinkedAppBatchContext(ctx context.Context, arg *RevokeLinkedApiAppBatchArg) (res *RevokeLinkedAppBatchResult, err error)
	// MemberSpaceLimitsExcludedUsersAddContext : Add users to member space
	// limits excluded users list.
	MemberSpaceLimitsExcludedUsersAddContext(ctx context.Context, arg *ExcludedUsersUpdateArg) (res *ExcludedUsersUpdateResult, err error)
	// MemberSpaceLimitsExcludedUsersListContext : List member space limits
	// excluded users.
	MemberSpaceLimitsExcludedUsersListContext(ctx context.Context, arg *ExcludedUsersListArg) (res *ExcludedUsersListResult, err error)
	// MemberSpaceLimitsExcludedUsersListContinueContext : Continue listing
	// member space limits excluded users.
	MemberSpaceLimitsExcludedUsersListContinueContext(ctx context.Context, arg *ExcludedUsersListContinueArg) (res *ExcludedUsersListResult, err error)
	// MemberSpaceLimitsExcludedUsersRemoveContext : Remove users from member
	// space limits excluded users list.
	MemberSpaceLimitsExcludedUsersRemoveContext(ctx context.Context, arg *ExcludedUsersUpdateArg) (res *ExcludedUsersUpdateResult, err error)
	// MemberSpaceLimitsGetCustomQuotaContext : Get users custom quota. A
	// maximum of 1000 members can be specified in a single call. Note: to apply
	// a custom space limit, a team admin needs to set a member space limit for
	// the team first. (the team admin can check the settings here:
	// https://www.dropbox.com/team/admin/settings/space).
	MemberSpaceLimitsGetCustomQuotaContext(ctx context.Context, arg *CustomQuotaUsersArg) (res []*CustomQuotaResult, err error)
	// MemberSpaceLimitsRemoveCustomQuotaContext : Remove users custom quota. A
	// maximum of 1000 members can be specified in a single call. Note: to apply
	// a custom space limit, a team admin needs to set a member space limit for
	// the team first. (the team admin can check the settings here:
	// https://www.dropbox.com/team/admin/settings/space).
	MemberSpaceLimitsRemoveCustomQuotaContext(ctx context.Context, arg *CustomQuotaUsersArg) (res []*RemoveCustomQuotaResult, err error)
	// MemberSpaceLimitsSetCustomQuotaContext : Set users custom quota. Custom
	// quota has to be at least 2GB. A maximum of 1000 members can be specified
	// in a single call. Note: to apply a custom space limit, a team admin needs
	// to set a member space limit for the team first. (the team admin can check
	// the settings here: https://www.dropbox.com/team/admin/settings/space).
	MemberSpaceLimitsSetCustomQuotaContext(ctx context.Context, arg *SetCustomQuotaArg) (res []*CustomQuotaResult, err error)
	// MembersAddContext : Adds members to a team. Permission : Team member
	// management A maximum of 20 members can be specified in a single call. If
	// no Dropbox account exists with the email address specified, a new Dropbox
	// account will be created with the given email address, and that account
	// will be invited to the team. If a personal Dropbox account exists with
	// the email address specified in the call, this call will create a
	// placeholder Dropbox account for the user on the team and send an email
	// inviting the user to migrate their existing personal account onto the
	// team. Team member management apps are required to set an initial
	// given_name and surname for a user to use in the team invitation and for
	// 'Perform as team member' actions taken on the user before they become
	// 'active'.
	MembersAddContext(ctx context.Context, arg *MembersAddArg) (res *MembersAddLaunch, err error)
	// MembersAddV2Context : Adds members to a team. Permission : Team member
	// management A maximum of 20 members can be specified in a single call. If
	// no Dropbox account exists with the email address specified, a new Dropbox
	// account will be created with the given email address, and that account
	// will be invited to the team. If a personal Dropbox account exists with
	// the email address specified in the call, this call will create a
	// placeholder Dropbox account for the user on the team and send an email
	// inviting the user to migrate their existing personal account onto the
	// team.
	MembersAddV2Context(ctx context.Context, arg *MembersAddV2Arg) (res *MembersAddLaunchV2Result, err error)
	// MembersAddJobStatusGetContext : Once an async_job_id is returned from
	// `membersAdd` , use this to poll the status of the asynchronous request.
	// Permission : Team member management.
	MembersAddJobStatusGetContext(ctx context.Context, arg *async.PollArg) (res *MembersAddJobStatus, err error)
	// MembersAddJobStatusGetV2Context : Once an async_job_id is returned from
	// `membersAdd` , use this to poll the status of the asynchronous request.
	// Permission : Team member management.
	MembersAddJobStatusGetV2Context(ctx context.Context, arg *async.PollArg) (res *MembersAddJobStatusV2Result, err error)
	// MembersBulkSuspendContext : Launch a bulk suspend job. The server
	// enforces a maximum of 500 members.
	MembersBulkSuspendContext(ctx context.Context, arg *BulkSuspendArg) (res *async.LaunchResultBase, err error)
	// MembersBulkSuspendJobStatusCheckContext : Poll a previously launched bulk
	// suspend job.
	MembersBulkSuspendJobStatusCheckContext(ctx context.Context, arg *async.PollArg) (res *BulkSuspendJobStatus, err error)
	// MembersDeleteFormerMemberFilesContext : Permanently delete the files of a
	// user who has been removed from the team. After permanent deletion, those
	// files will not be available to be transferred to another team member.
	// Permission : Team member management Exactly one of team_member_id, email,
	// or external_id must be provided to identify the user account.
	MembersDeleteFormerMemberFilesContext(ctx context.Context, arg *MembersFormerMemberArg) (err error)
	// MembersDeleteProfilePhotoContext : Deletes a team member's profile photo.
	// Permission : Team member management.
	MembersDeleteProfilePhotoContext(ctx context.Context, arg *MembersDeleteProfilePhotoArg) (res *TeamMemberInfo, err error)
	// MembersDeleteProfilePhotoV2Context : Deletes a team member's profile
	// photo. Permission : Team member management.
	MembersDeleteProfilePhotoV2Context(ctx context.Context, arg *MembersDeleteProfilePhotoArg) (res *TeamMemberInfoV2Result, err error)
	// MembersGetAvailableTeamMemberRolesContext : Get available TeamMemberRoles
	// for the connected team. To be used with `membersSetAdminPermissions`.
	// Permission : Team member management.
	MembersGetAvailableTeamMemberRolesContext(ctx context.Context) (res *MembersGetAvailableTeamMemberRolesResult, err error)
	// MembersGetInfoContext : Returns information about multiple team members.
	// Permission : Team information This endpoint will return
	// `MembersGetInfoItem.id_not_found`, for IDs (or emails) that cannot be
	// matched to a valid team member.
	MembersGetInfoContext(ctx context.Context, arg *MembersGetInfoArgs) (res []*MembersGetInfoItem, err error)
	// MembersGetInfoV2Context : Returns information about multiple team
	// members. Permission : Team information This endpoint will return
	// `MembersGetInfoItem.id_not_found`, for IDs (or emails) that cannot be
	// matched to a valid team member.
	MembersGetInfoV2Context(ctx context.Context, arg *MembersGetInfoV2Arg) (res *MembersGetInfoV2Result, err error)
	// MembersListContext : Lists members of a team. Permission : Team
	// information.
	MembersListContext(ctx context.Context, arg *MembersListArg) (res *MembersListResult, err error)
	// MembersListV2Context : Lists members of a team. Permission : Team
	// information.
	MembersListV2Context(ctx context.Context, arg *MembersListArg) (res *MembersListV2Result, err error)
	// MembersListContinueContext : Once a cursor has been retrieved from
	// `membersList`, use this to paginate through all team members. Permission
	// : Team information.
	MembersListContinueContext(ctx context.Context, arg *MembersListContinueArg) (res *MembersListResult, err error)
	// MembersListContinueV2Context : Once a cursor has been retrieved from
	// `membersList`, use this to paginate through all team members. Permission
	// : Team information.
	MembersListContinueV2Context(ctx context.Context, arg *MembersListContinueArg) (res *MembersListV2Result, err error)
	// MembersMoveFormerMemberFilesContext : Moves removed member's files to a
	// different member. This endpoint initiates an asynchronous job. To obtain
	// the final result of the job, the client should periodically poll
	// `membersMoveFormerMemberFilesJobStatusCheck`. Permission : Team member
	// management.
	MembersMoveFormerMemberFilesContext(ctx context.Context, arg *MembersDataTransferArg) (res *async.LaunchEmptyResult, err error)
	// MembersMoveFormerMemberFilesJobStatusCheckContext : Once an async_job_id
	// is returned from `membersMoveFormerMemberFiles` , use this to poll the
	// status of the asynchronous request. Permission : Team member management.
	MembersMoveFormerMemberFilesJobStatusCheckContext(ctx context.Context, arg *async.PollArg) (res *async.PollEmptyResult, err error)
	// MembersRecoverContext : Recover a deleted member. Permission : Team
	// member management Exactly one of team_member_id, email, or external_id
	// must be provided to identify the user account.
	MembersRecoverContext(ctx context.Context, arg *MembersRecoverArg) (err error)
	// MembersRemoveContext : Removes a member from a team. Permission : Team
	// member management Exactly one of team_member_id, email, or external_id
	// must be provided to identify the user account. Accounts can be recovered
	// via `membersRecover` for a 7 day period or until the account has been
	// permanently deleted or transferred to another account (whichever comes
	// first). Calling `membersAdd` while a user is still recoverable on your
	// team will return with `MemberAddResult.user_already_on_team`. Accounts
	// can have their files transferred via the admin console for a limited
	// time, based on the version history length associated with the team (180
	// days for most teams). Accounts can have their stacks transferred through
	// the admin console. This only transfers stacks that they have created.
	// This endpoint may initiate an asynchronous job. To obtain the final
	// result of the job, the client should periodically poll
	// `membersRemoveJobStatusGet`.
	MembersRemoveContext(ctx context.Context, arg *MembersRemoveArg) (res *async.LaunchEmptyResult, err error)
	// MembersRemoveJobStatusGetContext : Once an async_job_id is returned from
	// `membersRemove` , use this to poll the status of the asynchronous
	// request. Permission : Team member management.
	MembersRemoveJobStatusGetContext(ctx context.Context, arg *async.PollArg) (res *async.PollEmptyResult, err error)
	// MembersSecondaryEmailsAddContext : Add secondary emails to users.
	// Permission : Team member management. Emails that are on verified domains
	// will be verified automatically. For each email address not on a verified
	// domain a verification email will be sent.
	MembersSecondaryEmailsAddContext(ctx context.Context, arg *AddSecondaryEmailsArg) (res *AddSecondaryEmailsResult, err error)
	// MembersSecondaryEmailsDeleteContext : Delete secondary emails from users
	// Permission : Team member management. Users will be notified of deletions
	// of verified secondary emails at both the secondary email and their
	// primary email.
	MembersSecondaryEmailsDeleteContext(ctx context.Context, arg *DeleteSecondaryEmailsArg) (res *DeleteSecondaryEmailsResult, err error)
	// MembersSecondaryEmailsResendVerificationEmailsContext : Resend secondary
	// email verification emails. Permission : Team member management.
	MembersSecondaryEmailsResendVerificationEmailsContext(ctx context.Context, arg *ResendVerificationEmailArg) (res *ResendVerificationEmailResult, err error)
	// MembersSendWelcomeEmailContext : Sends welcome email to pending team
	// member. Permission : Team member management Exactly one of
	// team_member_id, email, or external_id must be provided to identify the
	// user account. No-op if team member is not pending.
	MembersSendWelcomeEmailContext(ctx context.Context, arg *UserSelectorArg) (err error)
	// MembersSetAdminPermissionsContext : Updates a team member's permissions.
	// Permission : Team member management.
	MembersSetAdminPermissionsContext(ctx context.Context, arg *MembersSetPermissionsArg) (res *MembersSetPermissionsResult, err error)
	// MembersSetAdminPermissionsV2Context : Updates a team member's
	// permissions. Permission : Team member management.
	MembersSetAdminPermissionsV2Context(ctx context.Context, arg *MembersSetPermissions2Arg) (res *MembersSetPermissions2Result, err error)
	// MembersSetProfileContext : Updates a team member's profile. Permission :
	// Team member management.
	MembersSetProfileContext(ctx context.Context, arg *MembersSetProfileArg) (res *TeamMemberInfo, err error)
	// MembersSetProfileV2Context : Updates a team member's profile. Permission
	// : Team member management.
	MembersSetProfileV2Context(ctx context.Context, arg *MembersSetProfileArg) (res *TeamMemberInfoV2Result, err error)
	// MembersSetProfilePhotoContext : Updates a team member's profile photo.
	// Permission : Team member management.
	MembersSetProfilePhotoContext(ctx context.Context, arg *MembersSetProfilePhotoArg) (res *TeamMemberInfo, err error)
	// MembersSetProfilePhotoV2Context : Updates a team member's profile photo.
	// Permission : Team member management.
	MembersSetProfilePhotoV2Context(ctx context.Context, arg *MembersSetProfilePhotoArg) (res *TeamMemberInfoV2Result, err error)
	// MembersSuspendContext : Suspend a member from a team. Permission : Team
	// member management Exactly one of team_member_id, email, or external_id
	// must be provided to identify the user account.
	MembersSuspendContext(ctx context.Context, arg *MembersDeactivateArg) (err error)
	// MembersUnsuspendContext : Unsuspend a member from a team. Permission :
	// Team member management Exactly one of team_member_id, email, or
	// external_id must be provided to identify the user account.
	MembersUnsuspendContext(ctx context.Context, arg *MembersUnsuspendArg) (err error)
	// NamespacesListContext : Returns a list of all team-accessible namespaces.
	// This list includes team folders, shared folders containing team members,
	// team members' home namespaces, and team members' app folders. Home
	// namespaces and app folders are always owned by this team or members of
	// the team, but shared folders may be owned by other users or other teams.
	// Duplicates may occur in the list.
	NamespacesListContext(ctx context.Context, arg *TeamNamespacesListArg) (res *TeamNamespacesListResult, err error)
	// NamespacesListContinueContext : Once a cursor has been retrieved from
	// `namespacesList`, use this to paginate through all team-accessible
	// namespaces. Duplicates may occur in the list.
	NamespacesListContinueContext(ctx context.Context, arg *TeamNamespacesListContinueArg) (res *TeamNamespacesListResult, err error)
	// PropertiesTemplateAddContext : Permission : Team member file access.
	// Deprecated:
	PropertiesTemplateAddContext(ctx context.Context, arg *file_properties.AddTemplateArg) (res *file_properties.AddTemplateResult, err error)
	// PropertiesTemplateGetContext : Permission : Team member file access. The
	// scope for the route is files.team_metadata.write.
	// Deprecated:
	PropertiesTemplateGetContext(ctx context.Context, arg *file_properties.GetTemplateArg) (res *file_properties.GetTemplateResult, err error)
	// ReportsGetActivityContext : Retrieves reporting data about a team's user
	// activity. Deprecated: Will be removed on July 1st 2021.
	// Deprecated:
	ReportsGetActivityContext(ctx context.Context, arg *DateRange) (res *GetActivityReport, err error)
	// ReportsGetDevicesContext : Retrieves reporting data about a team's linked
	// devices. Deprecated: Will be removed on July 1st 2021.
	// Deprecated:
	ReportsGetDevicesContext(ctx context.Context, arg *DateRange) (res *GetDevicesReport, err error)
	// ReportsGetMembershipContext : Retrieves reporting data about a team's
	// membership. Deprecated: Will be removed on July 1st 2021.
	// Deprecated:
	ReportsGetMembershipContext(ctx context.Context, arg *DateRange) (res *GetMembershipReport, err error)
	// ReportsGetStorageContext : Retrieves reporting data about a team's
	// storage usage. Deprecated: Will be removed on July 1st 2021.
	// Deprecated:
	ReportsGetStorageContext(ctx context.Context, arg *DateRange) (res *GetStorageReport, err error)
	// SharingAllowlistAddContext : Endpoint adds Approve List entries. Changes
	// are effective immediately. Changes are committed in transaction. In case
	// of single validation error - all entries are rejected. Valid domains
	// (RFC-1034/5) and emails (RFC-5322/822) are accepted. Added entries cannot
	// overflow limit of 10000 entries per team. Maximum 100 entries per call is
	// allowed.
	SharingAllowlistAddContext(ctx context.Context, arg *SharingAllowlistAddArgs) (res *SharingAllowlistAddResponse, err error)
	// SharingAllowlistListContext : Lists Approve List entries for given team,
	// from newest to oldest, returning up to `limit` entries at a time. If
	// there are more than `limit` entries associated with the current team,
	// more can be fetched by passing the returned `cursor` to
	// `sharingAllowlistListContinue`.
	SharingAllowlistListContext(ctx context.Context, arg *SharingAllowlistListArg) (res *SharingAllowlistListResponse, err error)
	// SharingAllowlistListContinueContext : Lists entries associated with given
	// team, starting from a the cursor. See `sharingAllowlistList`.
	SharingAllowlistListContinueContext(ctx context.Context, arg *SharingAllowlistListContinueArg) (res *SharingAllowlistListResponse, err error)
	// SharingAllowlistRemoveContext : Endpoint removes Approve List entries.
	// Changes are effective immediately. Changes are committed in transaction.
	// In case of single validation error - all entries are rejected. Valid
	// domains (RFC-1034/5) and emails (RFC-5322/822) are accepted. Entries
	// being removed have to be present on the list. Maximum 1000 entries per
	// call is allowed.
	SharingAllowlistRemoveContext(ctx context.Context, arg *SharingAllowlistRemoveArgs) (res *SharingAllowlistRemoveResponse, err error)
	// TeamFolderActivateContext : Sets an archived team folder's status to
	// active. Permission : Team member file access.
	TeamFolderActivateContext(ctx context.Context, arg *TeamFolderIdArg) (res *TeamFolderMetadata, err error)
	// TeamFolderArchiveContext : Sets an active team folder's status to
	// archived and removes all folder and file members. This endpoint cannot be
	// used for teams that have a shared team space. This route will either
	// finish synchronously, or return a job ID and do the async archive job in
	// background. Please use team_folder/archive/check to check the job status.
	// Permission : Team member file access.
	TeamFolderArchiveContext(ctx context.Context, arg *TeamFolderArchiveArg) (res *TeamFolderArchiveLaunch, err error)
	// TeamFolderArchiveCheckContext : Returns the status of an asynchronous job
	// for archiving a team folder. The job may show '.tag' as complete, but the
	// team folder could still be in the process of archiving (indicated by
	// `TeamFolderMetadata.status` with 'archive_in_progress'). To confirm that
	// the team folder is fully archived, check the field
	// `TeamFolderMetadata.status` in the response for the value 'archived'.
	// Permission : Team member file access.
	TeamFolderArchiveCheckContext(ctx context.Context, arg *async.PollArg) (res *TeamFolderArchiveJobStatus, err error)
	// TeamFolderCreateContext : Creates a new, active, team folder with no
	// members. This endpoint can only be used for teams that do not already
	// have a shared team space. Permission : Team member file access.
	TeamFolderCreateContext(ctx context.Context, arg *TeamFolderCreateArg) (res *TeamFolderMetadata, err error)
	// TeamFolderGetInfoContext : Retrieves metadata for team folders.
	// Permission : Team member file access.
	TeamFolderGetInfoContext(ctx context.Context, arg *TeamFolderIdListArg) (res []*TeamFolderGetInfoItem, err error)
	// TeamFolderListContext : Lists all team folders. Permission : Team member
	// file access.
	TeamFolderListContext(ctx context.Context, arg *TeamFolderListArg) (res *TeamFolderListResult, err error)
	// TeamFolderListContinueContext : Once a cursor has been retrieved from
	// `teamFolderList`, use this to paginate through all team folders.
	// Permission : Team member file access.
	TeamFolderListContinueContext(ctx context.Context, arg *TeamFolderListContinueArg) (res *TeamFolderListResult, err error)
	// TeamFolderPermanentlyDeleteContext : Permanently deletes an archived team
	// folder. This endpoint cannot be used for teams that have a shared team
	// space. Permission : Team member file access.
	TeamFolderPermanentlyDeleteContext(ctx context.Context, arg *TeamFolderIdArg) (err error)
	// TeamFolderRenameContext : Changes an active team folder's name.
	// Permission : Team member file access.
	TeamFolderRenameContext(ctx context.Context, arg *TeamFolderRenameArg) (res *TeamFolderMetadata, err error)
	// TeamFolderRestoreContext : Sets an inactive team folder's status to
	// active. Permission: Team member file access.
	TeamFolderRestoreContext(ctx context.Context, arg *TeamFolderIdArg) (res *TeamFolderMetadata, err error)
	// TeamFolderUpdateSyncSettingsContext : Updates the sync settings on a team
	// folder or its contents.  Use of this endpoint requires that the team has
	// team selective sync enabled.
	TeamFolderUpdateSyncSettingsContext(ctx context.Context, arg *TeamFolderUpdateSyncSettingsArg) (res *TeamFolderMetadata, err error)
	// TokenGetAuthenticatedAdminContext : Returns the member profile of the
	// admin who generated the team access token used to make the call.
	TokenGetAuthenticatedAdminContext(ctx context.Context) (res *TokenGetAuthenticatedAdminResult, err error)
}

type apiImpl dropbox.Context

// DevicesListMemberDevicesAPIError is an error-wrapper for the devices/list_member_devices route
type DevicesListMemberDevicesAPIError struct {
	dropbox.APIError
	EndpointError *ListMemberDevicesError `json:"error"`
}

// DevicesListMemberDevicesContext : List all device sessions of a team's
// member.
func (dbx *apiImpl) DevicesListMemberDevicesContext(ctx context.Context, arg *ListMemberDevicesArg) (res *ListMemberDevicesResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "devices/list_member_devices",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr DevicesListMemberDevicesAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) DevicesListMemberDevices(arg *ListMemberDevicesArg) (res *ListMemberDevicesResult, err error) {
	return dbx.DevicesListMemberDevicesContext(context.Background(), arg)
}

// DevicesListMembersDevicesAPIError is an error-wrapper for the devices/list_members_devices route
type DevicesListMembersDevicesAPIError struct {
	dropbox.APIError
	EndpointError *ListMembersDevicesError `json:"error"`
}

// DevicesListMembersDevicesContext : List all device sessions of a team.
// Permission : Team member file access.
func (dbx *apiImpl) DevicesListMembersDevicesContext(ctx context.Context, arg *ListMembersDevicesArg) (res *ListMembersDevicesResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "devices/list_members_devices",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr DevicesListMembersDevicesAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) DevicesListMembersDevices(arg *ListMembersDevicesArg) (res *ListMembersDevicesResult, err error) {
	return dbx.DevicesListMembersDevicesContext(context.Background(), arg)
}

// DevicesListTeamDevicesAPIError is an error-wrapper for the devices/list_team_devices route
type DevicesListTeamDevicesAPIError struct {
	dropbox.APIError
	EndpointError *ListTeamDevicesError `json:"error"`
}

// DevicesListTeamDevicesContext : List all device sessions of a team.
// Permission : Team member file access.
// Deprecated:
func (dbx *apiImpl) DevicesListTeamDevicesContext(ctx context.Context, arg *ListTeamDevicesArg) (res *ListTeamDevicesResult, err error) {
	log.Printf("WARNING: API `DevicesListTeamDevices` is deprecated")

	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "devices/list_team_devices",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr DevicesListTeamDevicesAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) DevicesListTeamDevices(arg *ListTeamDevicesArg) (res *ListTeamDevicesResult, err error) {
	return dbx.DevicesListTeamDevicesContext(context.Background(), arg)
}

// DevicesRevokeDeviceSessionAPIError is an error-wrapper for the devices/revoke_device_session route
type DevicesRevokeDeviceSessionAPIError struct {
	dropbox.APIError
	EndpointError *RevokeDeviceSessionError `json:"error"`
}

// DevicesRevokeDeviceSessionContext : Revoke a device session of a team's
// member.
func (dbx *apiImpl) DevicesRevokeDeviceSessionContext(ctx context.Context, arg *RevokeDeviceSessionArg) (err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "devices/revoke_device_session",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr DevicesRevokeDeviceSessionAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	_ = resp
	_ = respBody
	return
}

func (dbx *apiImpl) DevicesRevokeDeviceSession(arg *RevokeDeviceSessionArg) (err error) {
	return dbx.DevicesRevokeDeviceSessionContext(context.Background(), arg)
}

// DevicesRevokeDeviceSessionBatchAPIError is an error-wrapper for the devices/revoke_device_session_batch route
type DevicesRevokeDeviceSessionBatchAPIError struct {
	dropbox.APIError
	EndpointError *RevokeDeviceSessionBatchError `json:"error"`
}

// DevicesRevokeDeviceSessionBatchContext : Revoke a list of device sessions of
// team members.
func (dbx *apiImpl) DevicesRevokeDeviceSessionBatchContext(ctx context.Context, arg *RevokeDeviceSessionBatchArg) (res *RevokeDeviceSessionBatchResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "devices/revoke_device_session_batch",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr DevicesRevokeDeviceSessionBatchAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) DevicesRevokeDeviceSessionBatch(arg *RevokeDeviceSessionBatchArg) (res *RevokeDeviceSessionBatchResult, err error) {
	return dbx.DevicesRevokeDeviceSessionBatchContext(context.Background(), arg)
}

// FeaturesGetValuesAPIError is an error-wrapper for the features/get_values route
type FeaturesGetValuesAPIError struct {
	dropbox.APIError
	EndpointError *FeaturesGetValuesBatchError `json:"error"`
}

// FeaturesGetValuesContext : Get the values for one or more features. This
// route allows you to check your account's capability for what feature you can
// access or what value you have for certain features. Permission : Team
// information.
func (dbx *apiImpl) FeaturesGetValuesContext(ctx context.Context, arg *FeaturesGetValuesBatchArg) (res *FeaturesGetValuesBatchResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "features/get_values",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr FeaturesGetValuesAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) FeaturesGetValues(arg *FeaturesGetValuesBatchArg) (res *FeaturesGetValuesBatchResult, err error) {
	return dbx.FeaturesGetValuesContext(context.Background(), arg)
}

// GetInfoAPIError is an error-wrapper for the get_info route
type GetInfoAPIError struct {
	dropbox.APIError
	EndpointError struct{} `json:"error"`
}

// GetInfoContext : Retrieves information about a team.
func (dbx *apiImpl) GetInfoContext(ctx context.Context) (res *TeamGetInfoResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "get_info",
		Auth:         "team",
		Style:        "rpc",
		Arg:          nil,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr GetInfoAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) GetInfo() (res *TeamGetInfoResult, err error) {
	return dbx.GetInfoContext(context.Background())
}

// GroupsCreateAPIError is an error-wrapper for the groups/create route
type GroupsCreateAPIError struct {
	dropbox.APIError
	EndpointError *GroupCreateError `json:"error"`
}

// GroupsCreateContext : Creates a new, empty group, with a requested name.
// Permission : Team member management.
func (dbx *apiImpl) GroupsCreateContext(ctx context.Context, arg *GroupCreateArg) (res *GroupFullInfo, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "groups/create",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr GroupsCreateAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) GroupsCreate(arg *GroupCreateArg) (res *GroupFullInfo, err error) {
	return dbx.GroupsCreateContext(context.Background(), arg)
}

// GroupsDeleteAPIError is an error-wrapper for the groups/delete route
type GroupsDeleteAPIError struct {
	dropbox.APIError
	EndpointError *GroupDeleteError `json:"error"`
}

// GroupsDeleteContext : Deletes a group. The group is deleted immediately.
// However the revoking of group-owned resources may take additional time. Use
// the `groupsJobStatusGet` to determine whether this process has completed.
// Permission : Team member management.
func (dbx *apiImpl) GroupsDeleteContext(ctx context.Context, arg *GroupSelector) (res *async.LaunchEmptyResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "groups/delete",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr GroupsDeleteAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) GroupsDelete(arg *GroupSelector) (res *async.LaunchEmptyResult, err error) {
	return dbx.GroupsDeleteContext(context.Background(), arg)
}

// GroupsGetInfoAPIError is an error-wrapper for the groups/get_info route
type GroupsGetInfoAPIError struct {
	dropbox.APIError
	EndpointError *GroupsGetInfoError `json:"error"`
}

// GroupsGetInfoContext : Retrieves information about one or more groups. Note
// that the optional field `GroupFullInfo.members` is not returned for
// system-managed groups. Permission : Team Information.
func (dbx *apiImpl) GroupsGetInfoContext(ctx context.Context, arg *GroupsSelector) (res []*GroupsGetInfoItem, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "groups/get_info",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr GroupsGetInfoAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) GroupsGetInfo(arg *GroupsSelector) (res []*GroupsGetInfoItem, err error) {
	return dbx.GroupsGetInfoContext(context.Background(), arg)
}

// GroupsJobStatusGetAPIError is an error-wrapper for the groups/job_status/get route
type GroupsJobStatusGetAPIError struct {
	dropbox.APIError
	EndpointError *GroupsPollError `json:"error"`
}

// GroupsJobStatusGetContext : Once an async_job_id is returned from
// `groupsDelete`, `groupsMembersAdd` , or `groupsMembersRemove` use this method
// to poll the status of granting/revoking group members' access to group-owned
// resources. Permission : Team member management.
func (dbx *apiImpl) GroupsJobStatusGetContext(ctx context.Context, arg *async.PollArg) (res *async.PollEmptyResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "groups/job_status/get",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr GroupsJobStatusGetAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) GroupsJobStatusGet(arg *async.PollArg) (res *async.PollEmptyResult, err error) {
	return dbx.GroupsJobStatusGetContext(context.Background(), arg)
}

// GroupsListAPIError is an error-wrapper for the groups/list route
type GroupsListAPIError struct {
	dropbox.APIError
	EndpointError struct{} `json:"error"`
}

// GroupsListContext : Lists groups on a team. Permission : Team Information.
func (dbx *apiImpl) GroupsListContext(ctx context.Context, arg *GroupsListArg) (res *GroupsListResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "groups/list",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr GroupsListAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) GroupsList(arg *GroupsListArg) (res *GroupsListResult, err error) {
	return dbx.GroupsListContext(context.Background(), arg)
}

// GroupsListContinueAPIError is an error-wrapper for the groups/list/continue route
type GroupsListContinueAPIError struct {
	dropbox.APIError
	EndpointError *GroupsListContinueError `json:"error"`
}

// GroupsListContinueContext : Once a cursor has been retrieved from
// `groupsList`, use this to paginate through all groups. Permission : Team
// Information.
func (dbx *apiImpl) GroupsListContinueContext(ctx context.Context, arg *GroupsListContinueArg) (res *GroupsListResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "groups/list/continue",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr GroupsListContinueAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) GroupsListContinue(arg *GroupsListContinueArg) (res *GroupsListResult, err error) {
	return dbx.GroupsListContinueContext(context.Background(), arg)
}

// GroupsMembersAddAPIError is an error-wrapper for the groups/members/add route
type GroupsMembersAddAPIError struct {
	dropbox.APIError
	EndpointError *GroupMembersAddError `json:"error"`
}

// GroupsMembersAddContext : Adds members to a group. The members are added
// immediately. However the granting of group-owned resources may take
// additional time. Use the `groupsJobStatusGet` to determine whether this
// process has completed. Permission : Team member management.
func (dbx *apiImpl) GroupsMembersAddContext(ctx context.Context, arg *GroupMembersAddArg) (res *GroupMembersChangeResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "groups/members/add",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr GroupsMembersAddAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) GroupsMembersAdd(arg *GroupMembersAddArg) (res *GroupMembersChangeResult, err error) {
	return dbx.GroupsMembersAddContext(context.Background(), arg)
}

// GroupsMembersListAPIError is an error-wrapper for the groups/members/list route
type GroupsMembersListAPIError struct {
	dropbox.APIError
	EndpointError *GroupSelectorError `json:"error"`
}

// GroupsMembersListContext : Lists members of a group. Permission : Team
// Information.
func (dbx *apiImpl) GroupsMembersListContext(ctx context.Context, arg *GroupsMembersListArg) (res *GroupsMembersListResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "groups/members/list",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr GroupsMembersListAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) GroupsMembersList(arg *GroupsMembersListArg) (res *GroupsMembersListResult, err error) {
	return dbx.GroupsMembersListContext(context.Background(), arg)
}

// GroupsMembersListContinueAPIError is an error-wrapper for the groups/members/list/continue route
type GroupsMembersListContinueAPIError struct {
	dropbox.APIError
	EndpointError *GroupsMembersListContinueError `json:"error"`
}

// GroupsMembersListContinueContext : Once a cursor has been retrieved from
// `groupsMembersList`, use this to paginate through all members of the group.
// Permission : Team information.
func (dbx *apiImpl) GroupsMembersListContinueContext(ctx context.Context, arg *GroupsMembersListContinueArg) (res *GroupsMembersListResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "groups/members/list/continue",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr GroupsMembersListContinueAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) GroupsMembersListContinue(arg *GroupsMembersListContinueArg) (res *GroupsMembersListResult, err error) {
	return dbx.GroupsMembersListContinueContext(context.Background(), arg)
}

// GroupsMembersRemoveAPIError is an error-wrapper for the groups/members/remove route
type GroupsMembersRemoveAPIError struct {
	dropbox.APIError
	EndpointError *GroupMembersRemoveError `json:"error"`
}

// GroupsMembersRemoveContext : Removes members from a group. The members are
// removed immediately. However the revoking of group-owned resources may take
// additional time. Use the `groupsJobStatusGet` to determine whether this
// process has completed. This method permits removing the only owner of a
// group, even in cases where this is not possible via the web client.
// Permission : Team member management.
func (dbx *apiImpl) GroupsMembersRemoveContext(ctx context.Context, arg *GroupMembersRemoveArg) (res *GroupMembersChangeResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "groups/members/remove",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr GroupsMembersRemoveAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) GroupsMembersRemove(arg *GroupMembersRemoveArg) (res *GroupMembersChangeResult, err error) {
	return dbx.GroupsMembersRemoveContext(context.Background(), arg)
}

// GroupsMembersSetAccessTypeAPIError is an error-wrapper for the groups/members/set_access_type route
type GroupsMembersSetAccessTypeAPIError struct {
	dropbox.APIError
	EndpointError *GroupMemberSetAccessTypeError `json:"error"`
}

// GroupsMembersSetAccessTypeContext : Sets a member's access type in a group.
// Permission : Team member management.
func (dbx *apiImpl) GroupsMembersSetAccessTypeContext(ctx context.Context, arg *GroupMembersSetAccessTypeArg) (res []*GroupsGetInfoItem, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "groups/members/set_access_type",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr GroupsMembersSetAccessTypeAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) GroupsMembersSetAccessType(arg *GroupMembersSetAccessTypeArg) (res []*GroupsGetInfoItem, err error) {
	return dbx.GroupsMembersSetAccessTypeContext(context.Background(), arg)
}

// GroupsUpdateAPIError is an error-wrapper for the groups/update route
type GroupsUpdateAPIError struct {
	dropbox.APIError
	EndpointError *GroupUpdateError `json:"error"`
}

// GroupsUpdateContext : Updates a group's name and/or external ID. Permission :
// Team member management.
func (dbx *apiImpl) GroupsUpdateContext(ctx context.Context, arg *GroupUpdateArgs) (res *GroupFullInfo, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "groups/update",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr GroupsUpdateAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) GroupsUpdate(arg *GroupUpdateArgs) (res *GroupFullInfo, err error) {
	return dbx.GroupsUpdateContext(context.Background(), arg)
}

// LegalHoldsCreatePolicyAPIError is an error-wrapper for the legal_holds/create_policy route
type LegalHoldsCreatePolicyAPIError struct {
	dropbox.APIError
	EndpointError *LegalHoldsPolicyCreateError `json:"error"`
}

// LegalHoldsCreatePolicyContext : Creates new legal hold policy. Note: Legal
// Holds is a paid add-on. Not all teams have the feature. Permission : Team
// member file access.
func (dbx *apiImpl) LegalHoldsCreatePolicyContext(ctx context.Context, arg *LegalHoldsPolicyCreateArg) (res *LegalHoldPolicy, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "legal_holds/create_policy",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr LegalHoldsCreatePolicyAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) LegalHoldsCreatePolicy(arg *LegalHoldsPolicyCreateArg) (res *LegalHoldPolicy, err error) {
	return dbx.LegalHoldsCreatePolicyContext(context.Background(), arg)
}

// LegalHoldsGetPolicyAPIError is an error-wrapper for the legal_holds/get_policy route
type LegalHoldsGetPolicyAPIError struct {
	dropbox.APIError
	EndpointError *LegalHoldsGetPolicyError `json:"error"`
}

// LegalHoldsGetPolicyContext : Gets a legal hold by Id. Note: Legal Holds is a
// paid add-on. Not all teams have the feature. Permission : Team member file
// access.
func (dbx *apiImpl) LegalHoldsGetPolicyContext(ctx context.Context, arg *LegalHoldsGetPolicyArg) (res *LegalHoldPolicy, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "legal_holds/get_policy",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr LegalHoldsGetPolicyAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) LegalHoldsGetPolicy(arg *LegalHoldsGetPolicyArg) (res *LegalHoldPolicy, err error) {
	return dbx.LegalHoldsGetPolicyContext(context.Background(), arg)
}

// LegalHoldsListHeldRevisionsAPIError is an error-wrapper for the legal_holds/list_held_revisions route
type LegalHoldsListHeldRevisionsAPIError struct {
	dropbox.APIError
	EndpointError *LegalHoldsListHeldRevisionsError `json:"error"`
}

// LegalHoldsListHeldRevisionsContext : List the file metadata that's under the
// hold. Note: Legal Holds is a paid add-on. Not all teams have the feature.
// Permission : Team member file access.
func (dbx *apiImpl) LegalHoldsListHeldRevisionsContext(ctx context.Context, arg *LegalHoldsListHeldRevisionsArg) (res *LegalHoldsListHeldRevisionResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "legal_holds/list_held_revisions",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr LegalHoldsListHeldRevisionsAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) LegalHoldsListHeldRevisions(arg *LegalHoldsListHeldRevisionsArg) (res *LegalHoldsListHeldRevisionResult, err error) {
	return dbx.LegalHoldsListHeldRevisionsContext(context.Background(), arg)
}

// LegalHoldsListHeldRevisionsContinueAPIError is an error-wrapper for the legal_holds/list_held_revisions_continue route
type LegalHoldsListHeldRevisionsContinueAPIError struct {
	dropbox.APIError
	EndpointError *LegalHoldsListHeldRevisionsError `json:"error"`
}

// LegalHoldsListHeldRevisionsContinueContext : Continue listing the file
// metadata that's under the hold. Note: Legal Holds is a paid add-on. Not all
// teams have the feature. Permission : Team member file access.
func (dbx *apiImpl) LegalHoldsListHeldRevisionsContinueContext(ctx context.Context, arg *LegalHoldsListHeldRevisionsContinueArg) (res *LegalHoldsListHeldRevisionResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "legal_holds/list_held_revisions_continue",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr LegalHoldsListHeldRevisionsContinueAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) LegalHoldsListHeldRevisionsContinue(arg *LegalHoldsListHeldRevisionsContinueArg) (res *LegalHoldsListHeldRevisionResult, err error) {
	return dbx.LegalHoldsListHeldRevisionsContinueContext(context.Background(), arg)
}

// LegalHoldsListPoliciesAPIError is an error-wrapper for the legal_holds/list_policies route
type LegalHoldsListPoliciesAPIError struct {
	dropbox.APIError
	EndpointError *LegalHoldsListPoliciesError `json:"error"`
}

// LegalHoldsListPoliciesContext : Lists legal holds on a team. Note: Legal
// Holds is a paid add-on. Not all teams have the feature. Permission : Team
// member file access.
func (dbx *apiImpl) LegalHoldsListPoliciesContext(ctx context.Context, arg *LegalHoldsListPoliciesArg) (res *LegalHoldsListPoliciesResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "legal_holds/list_policies",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr LegalHoldsListPoliciesAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) LegalHoldsListPolicies(arg *LegalHoldsListPoliciesArg) (res *LegalHoldsListPoliciesResult, err error) {
	return dbx.LegalHoldsListPoliciesContext(context.Background(), arg)
}

// LegalHoldsReleasePolicyAPIError is an error-wrapper for the legal_holds/release_policy route
type LegalHoldsReleasePolicyAPIError struct {
	dropbox.APIError
	EndpointError *LegalHoldsPolicyReleaseError `json:"error"`
}

// LegalHoldsReleasePolicyContext : Releases a legal hold by Id. Note: Legal
// Holds is a paid add-on. Not all teams have the feature. Permission : Team
// member file access.
func (dbx *apiImpl) LegalHoldsReleasePolicyContext(ctx context.Context, arg *LegalHoldsPolicyReleaseArg) (err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "legal_holds/release_policy",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr LegalHoldsReleasePolicyAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	_ = resp
	_ = respBody
	return
}

func (dbx *apiImpl) LegalHoldsReleasePolicy(arg *LegalHoldsPolicyReleaseArg) (err error) {
	return dbx.LegalHoldsReleasePolicyContext(context.Background(), arg)
}

// LegalHoldsUpdatePolicyAPIError is an error-wrapper for the legal_holds/update_policy route
type LegalHoldsUpdatePolicyAPIError struct {
	dropbox.APIError
	EndpointError *LegalHoldsPolicyUpdateError `json:"error"`
}

// LegalHoldsUpdatePolicyContext : Updates a legal hold. Note: Legal Holds is a
// paid add-on. Not all teams have the feature. Permission : Team member file
// access.
func (dbx *apiImpl) LegalHoldsUpdatePolicyContext(ctx context.Context, arg *LegalHoldsPolicyUpdateArg) (res *LegalHoldPolicy, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "legal_holds/update_policy",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr LegalHoldsUpdatePolicyAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) LegalHoldsUpdatePolicy(arg *LegalHoldsPolicyUpdateArg) (res *LegalHoldPolicy, err error) {
	return dbx.LegalHoldsUpdatePolicyContext(context.Background(), arg)
}

// LinkedAppsListMemberLinkedAppsAPIError is an error-wrapper for the linked_apps/list_member_linked_apps route
type LinkedAppsListMemberLinkedAppsAPIError struct {
	dropbox.APIError
	EndpointError *ListMemberAppsError `json:"error"`
}

// LinkedAppsListMemberLinkedAppsContext : List all linked applications of the
// team member. Note, this endpoint does not list any team-linked applications.
func (dbx *apiImpl) LinkedAppsListMemberLinkedAppsContext(ctx context.Context, arg *ListMemberAppsArg) (res *ListMemberAppsResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "linked_apps/list_member_linked_apps",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr LinkedAppsListMemberLinkedAppsAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) LinkedAppsListMemberLinkedApps(arg *ListMemberAppsArg) (res *ListMemberAppsResult, err error) {
	return dbx.LinkedAppsListMemberLinkedAppsContext(context.Background(), arg)
}

// LinkedAppsListMembersLinkedAppsAPIError is an error-wrapper for the linked_apps/list_members_linked_apps route
type LinkedAppsListMembersLinkedAppsAPIError struct {
	dropbox.APIError
	EndpointError *ListMembersAppsError `json:"error"`
}

// LinkedAppsListMembersLinkedAppsContext : List all applications linked to the
// team members' accounts. Note, this endpoint does not list any team-linked
// applications.
func (dbx *apiImpl) LinkedAppsListMembersLinkedAppsContext(ctx context.Context, arg *ListMembersAppsArg) (res *ListMembersAppsResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "linked_apps/list_members_linked_apps",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr LinkedAppsListMembersLinkedAppsAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) LinkedAppsListMembersLinkedApps(arg *ListMembersAppsArg) (res *ListMembersAppsResult, err error) {
	return dbx.LinkedAppsListMembersLinkedAppsContext(context.Background(), arg)
}

// LinkedAppsListTeamLinkedAppsAPIError is an error-wrapper for the linked_apps/list_team_linked_apps route
type LinkedAppsListTeamLinkedAppsAPIError struct {
	dropbox.APIError
	EndpointError *ListTeamAppsError `json:"error"`
}

// LinkedAppsListTeamLinkedAppsContext : List all applications linked to the
// team members' accounts. Note, this endpoint doesn't list any team-linked
// applications.
// Deprecated:
func (dbx *apiImpl) LinkedAppsListTeamLinkedAppsContext(ctx context.Context, arg *ListTeamAppsArg) (res *ListTeamAppsResult, err error) {
	log.Printf("WARNING: API `LinkedAppsListTeamLinkedApps` is deprecated")

	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "linked_apps/list_team_linked_apps",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr LinkedAppsListTeamLinkedAppsAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) LinkedAppsListTeamLinkedApps(arg *ListTeamAppsArg) (res *ListTeamAppsResult, err error) {
	return dbx.LinkedAppsListTeamLinkedAppsContext(context.Background(), arg)
}

// LinkedAppsRevokeLinkedAppAPIError is an error-wrapper for the linked_apps/revoke_linked_app route
type LinkedAppsRevokeLinkedAppAPIError struct {
	dropbox.APIError
	EndpointError *RevokeLinkedAppError `json:"error"`
}

// LinkedAppsRevokeLinkedAppContext : Revoke a linked application of the team
// member.
func (dbx *apiImpl) LinkedAppsRevokeLinkedAppContext(ctx context.Context, arg *RevokeLinkedApiAppArg) (err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "linked_apps/revoke_linked_app",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr LinkedAppsRevokeLinkedAppAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	_ = resp
	_ = respBody
	return
}

func (dbx *apiImpl) LinkedAppsRevokeLinkedApp(arg *RevokeLinkedApiAppArg) (err error) {
	return dbx.LinkedAppsRevokeLinkedAppContext(context.Background(), arg)
}

// LinkedAppsRevokeLinkedAppBatchAPIError is an error-wrapper for the linked_apps/revoke_linked_app_batch route
type LinkedAppsRevokeLinkedAppBatchAPIError struct {
	dropbox.APIError
	EndpointError *RevokeLinkedAppBatchError `json:"error"`
}

// LinkedAppsRevokeLinkedAppBatchContext : Revoke a list of linked applications
// of the team members.
func (dbx *apiImpl) LinkedAppsRevokeLinkedAppBatchContext(ctx context.Context, arg *RevokeLinkedApiAppBatchArg) (res *RevokeLinkedAppBatchResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "linked_apps/revoke_linked_app_batch",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr LinkedAppsRevokeLinkedAppBatchAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) LinkedAppsRevokeLinkedAppBatch(arg *RevokeLinkedApiAppBatchArg) (res *RevokeLinkedAppBatchResult, err error) {
	return dbx.LinkedAppsRevokeLinkedAppBatchContext(context.Background(), arg)
}

// MemberSpaceLimitsExcludedUsersAddAPIError is an error-wrapper for the member_space_limits/excluded_users/add route
type MemberSpaceLimitsExcludedUsersAddAPIError struct {
	dropbox.APIError
	EndpointError *ExcludedUsersUpdateError `json:"error"`
}

// MemberSpaceLimitsExcludedUsersAddContext : Add users to member space limits
// excluded users list.
func (dbx *apiImpl) MemberSpaceLimitsExcludedUsersAddContext(ctx context.Context, arg *ExcludedUsersUpdateArg) (res *ExcludedUsersUpdateResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "member_space_limits/excluded_users/add",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr MemberSpaceLimitsExcludedUsersAddAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) MemberSpaceLimitsExcludedUsersAdd(arg *ExcludedUsersUpdateArg) (res *ExcludedUsersUpdateResult, err error) {
	return dbx.MemberSpaceLimitsExcludedUsersAddContext(context.Background(), arg)
}

// MemberSpaceLimitsExcludedUsersListAPIError is an error-wrapper for the member_space_limits/excluded_users/list route
type MemberSpaceLimitsExcludedUsersListAPIError struct {
	dropbox.APIError
	EndpointError *ExcludedUsersListError `json:"error"`
}

// MemberSpaceLimitsExcludedUsersListContext : List member space limits excluded
// users.
func (dbx *apiImpl) MemberSpaceLimitsExcludedUsersListContext(ctx context.Context, arg *ExcludedUsersListArg) (res *ExcludedUsersListResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "member_space_limits/excluded_users/list",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr MemberSpaceLimitsExcludedUsersListAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) MemberSpaceLimitsExcludedUsersList(arg *ExcludedUsersListArg) (res *ExcludedUsersListResult, err error) {
	return dbx.MemberSpaceLimitsExcludedUsersListContext(context.Background(), arg)
}

// MemberSpaceLimitsExcludedUsersListContinueAPIError is an error-wrapper for the member_space_limits/excluded_users/list/continue route
type MemberSpaceLimitsExcludedUsersListContinueAPIError struct {
	dropbox.APIError
	EndpointError *ExcludedUsersListContinueError `json:"error"`
}

// MemberSpaceLimitsExcludedUsersListContinueContext : Continue listing member
// space limits excluded users.
func (dbx *apiImpl) MemberSpaceLimitsExcludedUsersListContinueContext(ctx context.Context, arg *ExcludedUsersListContinueArg) (res *ExcludedUsersListResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "member_space_limits/excluded_users/list/continue",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr MemberSpaceLimitsExcludedUsersListContinueAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) MemberSpaceLimitsExcludedUsersListContinue(arg *ExcludedUsersListContinueArg) (res *ExcludedUsersListResult, err error) {
	return dbx.MemberSpaceLimitsExcludedUsersListContinueContext(context.Background(), arg)
}

// MemberSpaceLimitsExcludedUsersRemoveAPIError is an error-wrapper for the member_space_limits/excluded_users/remove route
type MemberSpaceLimitsExcludedUsersRemoveAPIError struct {
	dropbox.APIError
	EndpointError *ExcludedUsersUpdateError `json:"error"`
}

// MemberSpaceLimitsExcludedUsersRemoveContext : Remove users from member space
// limits excluded users list.
func (dbx *apiImpl) MemberSpaceLimitsExcludedUsersRemoveContext(ctx context.Context, arg *ExcludedUsersUpdateArg) (res *ExcludedUsersUpdateResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "member_space_limits/excluded_users/remove",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr MemberSpaceLimitsExcludedUsersRemoveAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) MemberSpaceLimitsExcludedUsersRemove(arg *ExcludedUsersUpdateArg) (res *ExcludedUsersUpdateResult, err error) {
	return dbx.MemberSpaceLimitsExcludedUsersRemoveContext(context.Background(), arg)
}

// MemberSpaceLimitsGetCustomQuotaAPIError is an error-wrapper for the member_space_limits/get_custom_quota route
type MemberSpaceLimitsGetCustomQuotaAPIError struct {
	dropbox.APIError
	EndpointError *CustomQuotaError `json:"error"`
}

// MemberSpaceLimitsGetCustomQuotaContext : Get users custom quota. A maximum of
// 1000 members can be specified in a single call. Note: to apply a custom space
// limit, a team admin needs to set a member space limit for the team first.
// (the team admin can check the settings here:
// https://www.dropbox.com/team/admin/settings/space).
func (dbx *apiImpl) MemberSpaceLimitsGetCustomQuotaContext(ctx context.Context, arg *CustomQuotaUsersArg) (res []*CustomQuotaResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "member_space_limits/get_custom_quota",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr MemberSpaceLimitsGetCustomQuotaAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) MemberSpaceLimitsGetCustomQuota(arg *CustomQuotaUsersArg) (res []*CustomQuotaResult, err error) {
	return dbx.MemberSpaceLimitsGetCustomQuotaContext(context.Background(), arg)
}

// MemberSpaceLimitsRemoveCustomQuotaAPIError is an error-wrapper for the member_space_limits/remove_custom_quota route
type MemberSpaceLimitsRemoveCustomQuotaAPIError struct {
	dropbox.APIError
	EndpointError *CustomQuotaError `json:"error"`
}

// MemberSpaceLimitsRemoveCustomQuotaContext : Remove users custom quota. A
// maximum of 1000 members can be specified in a single call. Note: to apply a
// custom space limit, a team admin needs to set a member space limit for the
// team first. (the team admin can check the settings here:
// https://www.dropbox.com/team/admin/settings/space).
func (dbx *apiImpl) MemberSpaceLimitsRemoveCustomQuotaContext(ctx context.Context, arg *CustomQuotaUsersArg) (res []*RemoveCustomQuotaResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "member_space_limits/remove_custom_quota",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr MemberSpaceLimitsRemoveCustomQuotaAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) MemberSpaceLimitsRemoveCustomQuota(arg *CustomQuotaUsersArg) (res []*RemoveCustomQuotaResult, err error) {
	return dbx.MemberSpaceLimitsRemoveCustomQuotaContext(context.Background(), arg)
}

// MemberSpaceLimitsSetCustomQuotaAPIError is an error-wrapper for the member_space_limits/set_custom_quota route
type MemberSpaceLimitsSetCustomQuotaAPIError struct {
	dropbox.APIError
	EndpointError *SetCustomQuotaError `json:"error"`
}

// MemberSpaceLimitsSetCustomQuotaContext : Set users custom quota. Custom quota
// has to be at least 2GB. A maximum of 1000 members can be specified in a
// single call. Note: to apply a custom space limit, a team admin needs to set a
// member space limit for the team first. (the team admin can check the settings
// here: https://www.dropbox.com/team/admin/settings/space).
func (dbx *apiImpl) MemberSpaceLimitsSetCustomQuotaContext(ctx context.Context, arg *SetCustomQuotaArg) (res []*CustomQuotaResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "member_space_limits/set_custom_quota",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr MemberSpaceLimitsSetCustomQuotaAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) MemberSpaceLimitsSetCustomQuota(arg *SetCustomQuotaArg) (res []*CustomQuotaResult, err error) {
	return dbx.MemberSpaceLimitsSetCustomQuotaContext(context.Background(), arg)
}

// MembersAddAPIError is an error-wrapper for the members/add route
type MembersAddAPIError struct {
	dropbox.APIError
	EndpointError struct{} `json:"error"`
}

// MembersAddContext : Adds members to a team. Permission : Team member
// management A maximum of 20 members can be specified in a single call. If no
// Dropbox account exists with the email address specified, a new Dropbox
// account will be created with the given email address, and that account will
// be invited to the team. If a personal Dropbox account exists with the email
// address specified in the call, this call will create a placeholder Dropbox
// account for the user on the team and send an email inviting the user to
// migrate their existing personal account onto the team. Team member management
// apps are required to set an initial given_name and surname for a user to use
// in the team invitation and for 'Perform as team member' actions taken on the
// user before they become 'active'.
func (dbx *apiImpl) MembersAddContext(ctx context.Context, arg *MembersAddArg) (res *MembersAddLaunch, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "members/add",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr MembersAddAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) MembersAdd(arg *MembersAddArg) (res *MembersAddLaunch, err error) {
	return dbx.MembersAddContext(context.Background(), arg)
}

// MembersAddV2APIError is an error-wrapper for the members/add_v2 route
type MembersAddV2APIError struct {
	dropbox.APIError
	EndpointError struct{} `json:"error"`
}

// MembersAddV2Context : Adds members to a team. Permission : Team member
// management A maximum of 20 members can be specified in a single call. If no
// Dropbox account exists with the email address specified, a new Dropbox
// account will be created with the given email address, and that account will
// be invited to the team. If a personal Dropbox account exists with the email
// address specified in the call, this call will create a placeholder Dropbox
// account for the user on the team and send an email inviting the user to
// migrate their existing personal account onto the team.
func (dbx *apiImpl) MembersAddV2Context(ctx context.Context, arg *MembersAddV2Arg) (res *MembersAddLaunchV2Result, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "members/add_v2",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr MembersAddV2APIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) MembersAddV2(arg *MembersAddV2Arg) (res *MembersAddLaunchV2Result, err error) {
	return dbx.MembersAddV2Context(context.Background(), arg)
}

// MembersAddJobStatusGetAPIError is an error-wrapper for the members/add/job_status/get route
type MembersAddJobStatusGetAPIError struct {
	dropbox.APIError
	EndpointError *async.PollError `json:"error"`
}

// MembersAddJobStatusGetContext : Once an async_job_id is returned from
// `membersAdd` , use this to poll the status of the asynchronous request.
// Permission : Team member management.
func (dbx *apiImpl) MembersAddJobStatusGetContext(ctx context.Context, arg *async.PollArg) (res *MembersAddJobStatus, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "members/add/job_status/get",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr MembersAddJobStatusGetAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) MembersAddJobStatusGet(arg *async.PollArg) (res *MembersAddJobStatus, err error) {
	return dbx.MembersAddJobStatusGetContext(context.Background(), arg)
}

// MembersAddJobStatusGetV2APIError is an error-wrapper for the members/add/job_status/get_v2 route
type MembersAddJobStatusGetV2APIError struct {
	dropbox.APIError
	EndpointError *async.PollError `json:"error"`
}

// MembersAddJobStatusGetV2Context : Once an async_job_id is returned from
// `membersAdd` , use this to poll the status of the asynchronous request.
// Permission : Team member management.
func (dbx *apiImpl) MembersAddJobStatusGetV2Context(ctx context.Context, arg *async.PollArg) (res *MembersAddJobStatusV2Result, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "members/add/job_status/get_v2",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr MembersAddJobStatusGetV2APIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) MembersAddJobStatusGetV2(arg *async.PollArg) (res *MembersAddJobStatusV2Result, err error) {
	return dbx.MembersAddJobStatusGetV2Context(context.Background(), arg)
}

// MembersBulkSuspendAPIError is an error-wrapper for the members/bulk_suspend route
type MembersBulkSuspendAPIError struct {
	dropbox.APIError
	EndpointError *BulkSuspendError `json:"error"`
}

// MembersBulkSuspendContext : Launch a bulk suspend job. The server enforces a
// maximum of 500 members.
func (dbx *apiImpl) MembersBulkSuspendContext(ctx context.Context, arg *BulkSuspendArg) (res *async.LaunchResultBase, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "members/bulk_suspend",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr MembersBulkSuspendAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) MembersBulkSuspend(arg *BulkSuspendArg) (res *async.LaunchResultBase, err error) {
	return dbx.MembersBulkSuspendContext(context.Background(), arg)
}

// MembersBulkSuspendJobStatusCheckAPIError is an error-wrapper for the members/bulk_suspend/job_status/check route
type MembersBulkSuspendJobStatusCheckAPIError struct {
	dropbox.APIError
	EndpointError *async.PollError `json:"error"`
}

// MembersBulkSuspendJobStatusCheckContext : Poll a previously launched bulk
// suspend job.
func (dbx *apiImpl) MembersBulkSuspendJobStatusCheckContext(ctx context.Context, arg *async.PollArg) (res *BulkSuspendJobStatus, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "members/bulk_suspend/job_status/check",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr MembersBulkSuspendJobStatusCheckAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) MembersBulkSuspendJobStatusCheck(arg *async.PollArg) (res *BulkSuspendJobStatus, err error) {
	return dbx.MembersBulkSuspendJobStatusCheckContext(context.Background(), arg)
}

// MembersDeleteFormerMemberFilesAPIError is an error-wrapper for the members/delete_former_member_files route
type MembersDeleteFormerMemberFilesAPIError struct {
	dropbox.APIError
	EndpointError *MembersDeleteFormerMemberFilesError `json:"error"`
}

// MembersDeleteFormerMemberFilesContext : Permanently delete the files of a
// user who has been removed from the team. After permanent deletion, those
// files will not be available to be transferred to another team member.
// Permission : Team member management Exactly one of team_member_id, email, or
// external_id must be provided to identify the user account.
func (dbx *apiImpl) MembersDeleteFormerMemberFilesContext(ctx context.Context, arg *MembersFormerMemberArg) (err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "members/delete_former_member_files",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr MembersDeleteFormerMemberFilesAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	_ = resp
	_ = respBody
	return
}

func (dbx *apiImpl) MembersDeleteFormerMemberFiles(arg *MembersFormerMemberArg) (err error) {
	return dbx.MembersDeleteFormerMemberFilesContext(context.Background(), arg)
}

// MembersDeleteProfilePhotoAPIError is an error-wrapper for the members/delete_profile_photo route
type MembersDeleteProfilePhotoAPIError struct {
	dropbox.APIError
	EndpointError *MembersDeleteProfilePhotoError `json:"error"`
}

// MembersDeleteProfilePhotoContext : Deletes a team member's profile photo.
// Permission : Team member management.
func (dbx *apiImpl) MembersDeleteProfilePhotoContext(ctx context.Context, arg *MembersDeleteProfilePhotoArg) (res *TeamMemberInfo, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "members/delete_profile_photo",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr MembersDeleteProfilePhotoAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) MembersDeleteProfilePhoto(arg *MembersDeleteProfilePhotoArg) (res *TeamMemberInfo, err error) {
	return dbx.MembersDeleteProfilePhotoContext(context.Background(), arg)
}

// MembersDeleteProfilePhotoV2APIError is an error-wrapper for the members/delete_profile_photo_v2 route
type MembersDeleteProfilePhotoV2APIError struct {
	dropbox.APIError
	EndpointError *MembersDeleteProfilePhotoError `json:"error"`
}

// MembersDeleteProfilePhotoV2Context : Deletes a team member's profile photo.
// Permission : Team member management.
func (dbx *apiImpl) MembersDeleteProfilePhotoV2Context(ctx context.Context, arg *MembersDeleteProfilePhotoArg) (res *TeamMemberInfoV2Result, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "members/delete_profile_photo_v2",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr MembersDeleteProfilePhotoV2APIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) MembersDeleteProfilePhotoV2(arg *MembersDeleteProfilePhotoArg) (res *TeamMemberInfoV2Result, err error) {
	return dbx.MembersDeleteProfilePhotoV2Context(context.Background(), arg)
}

// MembersGetAvailableTeamMemberRolesAPIError is an error-wrapper for the members/get_available_team_member_roles route
type MembersGetAvailableTeamMemberRolesAPIError struct {
	dropbox.APIError
	EndpointError struct{} `json:"error"`
}

// MembersGetAvailableTeamMemberRolesContext : Get available TeamMemberRoles for
// the connected team. To be used with `membersSetAdminPermissions`. Permission
// : Team member management.
func (dbx *apiImpl) MembersGetAvailableTeamMemberRolesContext(ctx context.Context) (res *MembersGetAvailableTeamMemberRolesResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "members/get_available_team_member_roles",
		Auth:         "team",
		Style:        "rpc",
		Arg:          nil,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr MembersGetAvailableTeamMemberRolesAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) MembersGetAvailableTeamMemberRoles() (res *MembersGetAvailableTeamMemberRolesResult, err error) {
	return dbx.MembersGetAvailableTeamMemberRolesContext(context.Background())
}

// MembersGetInfoAPIError is an error-wrapper for the members/get_info route
type MembersGetInfoAPIError struct {
	dropbox.APIError
	EndpointError *MembersGetInfoError `json:"error"`
}

// MembersGetInfoContext : Returns information about multiple team members.
// Permission : Team information This endpoint will return
// `MembersGetInfoItem.id_not_found`, for IDs (or emails) that cannot be matched
// to a valid team member.
func (dbx *apiImpl) MembersGetInfoContext(ctx context.Context, arg *MembersGetInfoArgs) (res []*MembersGetInfoItem, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "members/get_info",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr MembersGetInfoAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) MembersGetInfo(arg *MembersGetInfoArgs) (res []*MembersGetInfoItem, err error) {
	return dbx.MembersGetInfoContext(context.Background(), arg)
}

// MembersGetInfoV2APIError is an error-wrapper for the members/get_info_v2 route
type MembersGetInfoV2APIError struct {
	dropbox.APIError
	EndpointError *MembersGetInfoError `json:"error"`
}

// MembersGetInfoV2Context : Returns information about multiple team members.
// Permission : Team information This endpoint will return
// `MembersGetInfoItem.id_not_found`, for IDs (or emails) that cannot be matched
// to a valid team member.
func (dbx *apiImpl) MembersGetInfoV2Context(ctx context.Context, arg *MembersGetInfoV2Arg) (res *MembersGetInfoV2Result, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "members/get_info_v2",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr MembersGetInfoV2APIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) MembersGetInfoV2(arg *MembersGetInfoV2Arg) (res *MembersGetInfoV2Result, err error) {
	return dbx.MembersGetInfoV2Context(context.Background(), arg)
}

// MembersListAPIError is an error-wrapper for the members/list route
type MembersListAPIError struct {
	dropbox.APIError
	EndpointError *MembersListError `json:"error"`
}

// MembersListContext : Lists members of a team. Permission : Team information.
func (dbx *apiImpl) MembersListContext(ctx context.Context, arg *MembersListArg) (res *MembersListResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "members/list",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr MembersListAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) MembersList(arg *MembersListArg) (res *MembersListResult, err error) {
	return dbx.MembersListContext(context.Background(), arg)
}

// MembersListV2APIError is an error-wrapper for the members/list_v2 route
type MembersListV2APIError struct {
	dropbox.APIError
	EndpointError *MembersListError `json:"error"`
}

// MembersListV2Context : Lists members of a team. Permission : Team
// information.
func (dbx *apiImpl) MembersListV2Context(ctx context.Context, arg *MembersListArg) (res *MembersListV2Result, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "members/list_v2",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr MembersListV2APIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) MembersListV2(arg *MembersListArg) (res *MembersListV2Result, err error) {
	return dbx.MembersListV2Context(context.Background(), arg)
}

// MembersListContinueAPIError is an error-wrapper for the members/list/continue route
type MembersListContinueAPIError struct {
	dropbox.APIError
	EndpointError *MembersListContinueError `json:"error"`
}

// MembersListContinueContext : Once a cursor has been retrieved from
// `membersList`, use this to paginate through all team members. Permission :
// Team information.
func (dbx *apiImpl) MembersListContinueContext(ctx context.Context, arg *MembersListContinueArg) (res *MembersListResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "members/list/continue",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr MembersListContinueAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) MembersListContinue(arg *MembersListContinueArg) (res *MembersListResult, err error) {
	return dbx.MembersListContinueContext(context.Background(), arg)
}

// MembersListContinueV2APIError is an error-wrapper for the members/list/continue_v2 route
type MembersListContinueV2APIError struct {
	dropbox.APIError
	EndpointError *MembersListContinueError `json:"error"`
}

// MembersListContinueV2Context : Once a cursor has been retrieved from
// `membersList`, use this to paginate through all team members. Permission :
// Team information.
func (dbx *apiImpl) MembersListContinueV2Context(ctx context.Context, arg *MembersListContinueArg) (res *MembersListV2Result, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "members/list/continue_v2",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr MembersListContinueV2APIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) MembersListContinueV2(arg *MembersListContinueArg) (res *MembersListV2Result, err error) {
	return dbx.MembersListContinueV2Context(context.Background(), arg)
}

// MembersMoveFormerMemberFilesAPIError is an error-wrapper for the members/move_former_member_files route
type MembersMoveFormerMemberFilesAPIError struct {
	dropbox.APIError
	EndpointError *MembersTransferFormerMembersFilesError `json:"error"`
}

// MembersMoveFormerMemberFilesContext : Moves removed member's files to a
// different member. This endpoint initiates an asynchronous job. To obtain the
// final result of the job, the client should periodically poll
// `membersMoveFormerMemberFilesJobStatusCheck`. Permission : Team member
// management.
func (dbx *apiImpl) MembersMoveFormerMemberFilesContext(ctx context.Context, arg *MembersDataTransferArg) (res *async.LaunchEmptyResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "members/move_former_member_files",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr MembersMoveFormerMemberFilesAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) MembersMoveFormerMemberFiles(arg *MembersDataTransferArg) (res *async.LaunchEmptyResult, err error) {
	return dbx.MembersMoveFormerMemberFilesContext(context.Background(), arg)
}

// MembersMoveFormerMemberFilesJobStatusCheckAPIError is an error-wrapper for the members/move_former_member_files/job_status/check route
type MembersMoveFormerMemberFilesJobStatusCheckAPIError struct {
	dropbox.APIError
	EndpointError *async.PollError `json:"error"`
}

// MembersMoveFormerMemberFilesJobStatusCheckContext : Once an async_job_id is
// returned from `membersMoveFormerMemberFiles` , use this to poll the status of
// the asynchronous request. Permission : Team member management.
func (dbx *apiImpl) MembersMoveFormerMemberFilesJobStatusCheckContext(ctx context.Context, arg *async.PollArg) (res *async.PollEmptyResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "members/move_former_member_files/job_status/check",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr MembersMoveFormerMemberFilesJobStatusCheckAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) MembersMoveFormerMemberFilesJobStatusCheck(arg *async.PollArg) (res *async.PollEmptyResult, err error) {
	return dbx.MembersMoveFormerMemberFilesJobStatusCheckContext(context.Background(), arg)
}

// MembersRecoverAPIError is an error-wrapper for the members/recover route
type MembersRecoverAPIError struct {
	dropbox.APIError
	EndpointError *MembersRecoverError `json:"error"`
}

// MembersRecoverContext : Recover a deleted member. Permission : Team member
// management Exactly one of team_member_id, email, or external_id must be
// provided to identify the user account.
func (dbx *apiImpl) MembersRecoverContext(ctx context.Context, arg *MembersRecoverArg) (err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "members/recover",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr MembersRecoverAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	_ = resp
	_ = respBody
	return
}

func (dbx *apiImpl) MembersRecover(arg *MembersRecoverArg) (err error) {
	return dbx.MembersRecoverContext(context.Background(), arg)
}

// MembersRemoveAPIError is an error-wrapper for the members/remove route
type MembersRemoveAPIError struct {
	dropbox.APIError
	EndpointError *MembersRemoveError `json:"error"`
}

// MembersRemoveContext : Removes a member from a team. Permission : Team member
// management Exactly one of team_member_id, email, or external_id must be
// provided to identify the user account. Accounts can be recovered via
// `membersRecover` for a 7 day period or until the account has been permanently
// deleted or transferred to another account (whichever comes first). Calling
// `membersAdd` while a user is still recoverable on your team will return with
// `MemberAddResult.user_already_on_team`. Accounts can have their files
// transferred via the admin console for a limited time, based on the version
// history length associated with the team (180 days for most teams). Accounts
// can have their stacks transferred through the admin console. This only
// transfers stacks that they have created. This endpoint may initiate an
// asynchronous job. To obtain the final result of the job, the client should
// periodically poll `membersRemoveJobStatusGet`.
func (dbx *apiImpl) MembersRemoveContext(ctx context.Context, arg *MembersRemoveArg) (res *async.LaunchEmptyResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "members/remove",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr MembersRemoveAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) MembersRemove(arg *MembersRemoveArg) (res *async.LaunchEmptyResult, err error) {
	return dbx.MembersRemoveContext(context.Background(), arg)
}

// MembersRemoveJobStatusGetAPIError is an error-wrapper for the members/remove/job_status/get route
type MembersRemoveJobStatusGetAPIError struct {
	dropbox.APIError
	EndpointError *async.PollError `json:"error"`
}

// MembersRemoveJobStatusGetContext : Once an async_job_id is returned from
// `membersRemove` , use this to poll the status of the asynchronous request.
// Permission : Team member management.
func (dbx *apiImpl) MembersRemoveJobStatusGetContext(ctx context.Context, arg *async.PollArg) (res *async.PollEmptyResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "members/remove/job_status/get",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr MembersRemoveJobStatusGetAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) MembersRemoveJobStatusGet(arg *async.PollArg) (res *async.PollEmptyResult, err error) {
	return dbx.MembersRemoveJobStatusGetContext(context.Background(), arg)
}

// MembersSecondaryEmailsAddAPIError is an error-wrapper for the members/secondary_emails/add route
type MembersSecondaryEmailsAddAPIError struct {
	dropbox.APIError
	EndpointError *AddSecondaryEmailsError `json:"error"`
}

// MembersSecondaryEmailsAddContext : Add secondary emails to users. Permission
// : Team member management. Emails that are on verified domains will be
// verified automatically. For each email address not on a verified domain a
// verification email will be sent.
func (dbx *apiImpl) MembersSecondaryEmailsAddContext(ctx context.Context, arg *AddSecondaryEmailsArg) (res *AddSecondaryEmailsResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "members/secondary_emails/add",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr MembersSecondaryEmailsAddAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) MembersSecondaryEmailsAdd(arg *AddSecondaryEmailsArg) (res *AddSecondaryEmailsResult, err error) {
	return dbx.MembersSecondaryEmailsAddContext(context.Background(), arg)
}

// MembersSecondaryEmailsDeleteAPIError is an error-wrapper for the members/secondary_emails/delete route
type MembersSecondaryEmailsDeleteAPIError struct {
	dropbox.APIError
	EndpointError struct{} `json:"error"`
}

// MembersSecondaryEmailsDeleteContext : Delete secondary emails from users
// Permission : Team member management. Users will be notified of deletions of
// verified secondary emails at both the secondary email and their primary
// email.
func (dbx *apiImpl) MembersSecondaryEmailsDeleteContext(ctx context.Context, arg *DeleteSecondaryEmailsArg) (res *DeleteSecondaryEmailsResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "members/secondary_emails/delete",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr MembersSecondaryEmailsDeleteAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) MembersSecondaryEmailsDelete(arg *DeleteSecondaryEmailsArg) (res *DeleteSecondaryEmailsResult, err error) {
	return dbx.MembersSecondaryEmailsDeleteContext(context.Background(), arg)
}

// MembersSecondaryEmailsResendVerificationEmailsAPIError is an error-wrapper for the members/secondary_emails/resend_verification_emails route
type MembersSecondaryEmailsResendVerificationEmailsAPIError struct {
	dropbox.APIError
	EndpointError struct{} `json:"error"`
}

// MembersSecondaryEmailsResendVerificationEmailsContext : Resend secondary
// email verification emails. Permission : Team member management.
func (dbx *apiImpl) MembersSecondaryEmailsResendVerificationEmailsContext(ctx context.Context, arg *ResendVerificationEmailArg) (res *ResendVerificationEmailResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "members/secondary_emails/resend_verification_emails",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr MembersSecondaryEmailsResendVerificationEmailsAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) MembersSecondaryEmailsResendVerificationEmails(arg *ResendVerificationEmailArg) (res *ResendVerificationEmailResult, err error) {
	return dbx.MembersSecondaryEmailsResendVerificationEmailsContext(context.Background(), arg)
}

// MembersSendWelcomeEmailAPIError is an error-wrapper for the members/send_welcome_email route
type MembersSendWelcomeEmailAPIError struct {
	dropbox.APIError
	EndpointError *MembersSendWelcomeError `json:"error"`
}

// MembersSendWelcomeEmailContext : Sends welcome email to pending team member.
// Permission : Team member management Exactly one of team_member_id, email, or
// external_id must be provided to identify the user account. No-op if team
// member is not pending.
func (dbx *apiImpl) MembersSendWelcomeEmailContext(ctx context.Context, arg *UserSelectorArg) (err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "members/send_welcome_email",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr MembersSendWelcomeEmailAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	_ = resp
	_ = respBody
	return
}

func (dbx *apiImpl) MembersSendWelcomeEmail(arg *UserSelectorArg) (err error) {
	return dbx.MembersSendWelcomeEmailContext(context.Background(), arg)
}

// MembersSetAdminPermissionsAPIError is an error-wrapper for the members/set_admin_permissions route
type MembersSetAdminPermissionsAPIError struct {
	dropbox.APIError
	EndpointError *MembersSetPermissionsError `json:"error"`
}

// MembersSetAdminPermissionsContext : Updates a team member's permissions.
// Permission : Team member management.
func (dbx *apiImpl) MembersSetAdminPermissionsContext(ctx context.Context, arg *MembersSetPermissionsArg) (res *MembersSetPermissionsResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "members/set_admin_permissions",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr MembersSetAdminPermissionsAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) MembersSetAdminPermissions(arg *MembersSetPermissionsArg) (res *MembersSetPermissionsResult, err error) {
	return dbx.MembersSetAdminPermissionsContext(context.Background(), arg)
}

// MembersSetAdminPermissionsV2APIError is an error-wrapper for the members/set_admin_permissions_v2 route
type MembersSetAdminPermissionsV2APIError struct {
	dropbox.APIError
	EndpointError *MembersSetPermissions2Error `json:"error"`
}

// MembersSetAdminPermissionsV2Context : Updates a team member's permissions.
// Permission : Team member management.
func (dbx *apiImpl) MembersSetAdminPermissionsV2Context(ctx context.Context, arg *MembersSetPermissions2Arg) (res *MembersSetPermissions2Result, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "members/set_admin_permissions_v2",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr MembersSetAdminPermissionsV2APIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) MembersSetAdminPermissionsV2(arg *MembersSetPermissions2Arg) (res *MembersSetPermissions2Result, err error) {
	return dbx.MembersSetAdminPermissionsV2Context(context.Background(), arg)
}

// MembersSetProfileAPIError is an error-wrapper for the members/set_profile route
type MembersSetProfileAPIError struct {
	dropbox.APIError
	EndpointError *MembersSetProfileError `json:"error"`
}

// MembersSetProfileContext : Updates a team member's profile. Permission : Team
// member management.
func (dbx *apiImpl) MembersSetProfileContext(ctx context.Context, arg *MembersSetProfileArg) (res *TeamMemberInfo, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "members/set_profile",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr MembersSetProfileAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) MembersSetProfile(arg *MembersSetProfileArg) (res *TeamMemberInfo, err error) {
	return dbx.MembersSetProfileContext(context.Background(), arg)
}

// MembersSetProfileV2APIError is an error-wrapper for the members/set_profile_v2 route
type MembersSetProfileV2APIError struct {
	dropbox.APIError
	EndpointError *MembersSetProfileError `json:"error"`
}

// MembersSetProfileV2Context : Updates a team member's profile. Permission :
// Team member management.
func (dbx *apiImpl) MembersSetProfileV2Context(ctx context.Context, arg *MembersSetProfileArg) (res *TeamMemberInfoV2Result, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "members/set_profile_v2",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr MembersSetProfileV2APIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) MembersSetProfileV2(arg *MembersSetProfileArg) (res *TeamMemberInfoV2Result, err error) {
	return dbx.MembersSetProfileV2Context(context.Background(), arg)
}

// MembersSetProfilePhotoAPIError is an error-wrapper for the members/set_profile_photo route
type MembersSetProfilePhotoAPIError struct {
	dropbox.APIError
	EndpointError *MembersSetProfilePhotoError `json:"error"`
}

// MembersSetProfilePhotoContext : Updates a team member's profile photo.
// Permission : Team member management.
func (dbx *apiImpl) MembersSetProfilePhotoContext(ctx context.Context, arg *MembersSetProfilePhotoArg) (res *TeamMemberInfo, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "members/set_profile_photo",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr MembersSetProfilePhotoAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) MembersSetProfilePhoto(arg *MembersSetProfilePhotoArg) (res *TeamMemberInfo, err error) {
	return dbx.MembersSetProfilePhotoContext(context.Background(), arg)
}

// MembersSetProfilePhotoV2APIError is an error-wrapper for the members/set_profile_photo_v2 route
type MembersSetProfilePhotoV2APIError struct {
	dropbox.APIError
	EndpointError *MembersSetProfilePhotoError `json:"error"`
}

// MembersSetProfilePhotoV2Context : Updates a team member's profile photo.
// Permission : Team member management.
func (dbx *apiImpl) MembersSetProfilePhotoV2Context(ctx context.Context, arg *MembersSetProfilePhotoArg) (res *TeamMemberInfoV2Result, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "members/set_profile_photo_v2",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr MembersSetProfilePhotoV2APIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) MembersSetProfilePhotoV2(arg *MembersSetProfilePhotoArg) (res *TeamMemberInfoV2Result, err error) {
	return dbx.MembersSetProfilePhotoV2Context(context.Background(), arg)
}

// MembersSuspendAPIError is an error-wrapper for the members/suspend route
type MembersSuspendAPIError struct {
	dropbox.APIError
	EndpointError *MembersSuspendError `json:"error"`
}

// MembersSuspendContext : Suspend a member from a team. Permission : Team
// member management Exactly one of team_member_id, email, or external_id must
// be provided to identify the user account.
func (dbx *apiImpl) MembersSuspendContext(ctx context.Context, arg *MembersDeactivateArg) (err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "members/suspend",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr MembersSuspendAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	_ = resp
	_ = respBody
	return
}

func (dbx *apiImpl) MembersSuspend(arg *MembersDeactivateArg) (err error) {
	return dbx.MembersSuspendContext(context.Background(), arg)
}

// MembersUnsuspendAPIError is an error-wrapper for the members/unsuspend route
type MembersUnsuspendAPIError struct {
	dropbox.APIError
	EndpointError *MembersUnsuspendError `json:"error"`
}

// MembersUnsuspendContext : Unsuspend a member from a team. Permission : Team
// member management Exactly one of team_member_id, email, or external_id must
// be provided to identify the user account.
func (dbx *apiImpl) MembersUnsuspendContext(ctx context.Context, arg *MembersUnsuspendArg) (err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "members/unsuspend",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr MembersUnsuspendAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	_ = resp
	_ = respBody
	return
}

func (dbx *apiImpl) MembersUnsuspend(arg *MembersUnsuspendArg) (err error) {
	return dbx.MembersUnsuspendContext(context.Background(), arg)
}

// NamespacesListAPIError is an error-wrapper for the namespaces/list route
type NamespacesListAPIError struct {
	dropbox.APIError
	EndpointError *TeamNamespacesListError `json:"error"`
}

// NamespacesListContext : Returns a list of all team-accessible namespaces.
// This list includes team folders, shared folders containing team members, team
// members' home namespaces, and team members' app folders. Home namespaces and
// app folders are always owned by this team or members of the team, but shared
// folders may be owned by other users or other teams. Duplicates may occur in
// the list.
func (dbx *apiImpl) NamespacesListContext(ctx context.Context, arg *TeamNamespacesListArg) (res *TeamNamespacesListResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "namespaces/list",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr NamespacesListAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) NamespacesList(arg *TeamNamespacesListArg) (res *TeamNamespacesListResult, err error) {
	return dbx.NamespacesListContext(context.Background(), arg)
}

// NamespacesListContinueAPIError is an error-wrapper for the namespaces/list/continue route
type NamespacesListContinueAPIError struct {
	dropbox.APIError
	EndpointError *TeamNamespacesListContinueError `json:"error"`
}

// NamespacesListContinueContext : Once a cursor has been retrieved from
// `namespacesList`, use this to paginate through all team-accessible
// namespaces. Duplicates may occur in the list.
func (dbx *apiImpl) NamespacesListContinueContext(ctx context.Context, arg *TeamNamespacesListContinueArg) (res *TeamNamespacesListResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "namespaces/list/continue",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr NamespacesListContinueAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) NamespacesListContinue(arg *TeamNamespacesListContinueArg) (res *TeamNamespacesListResult, err error) {
	return dbx.NamespacesListContinueContext(context.Background(), arg)
}

// PropertiesTemplateAddAPIError is an error-wrapper for the properties/template/add route
type PropertiesTemplateAddAPIError struct {
	dropbox.APIError
	EndpointError *file_properties.ModifyTemplateError `json:"error"`
}

// PropertiesTemplateAddContext : Permission : Team member file access.
// Deprecated:
func (dbx *apiImpl) PropertiesTemplateAddContext(ctx context.Context, arg *file_properties.AddTemplateArg) (res *file_properties.AddTemplateResult, err error) {
	log.Printf("WARNING: API `PropertiesTemplateAdd` is deprecated")

	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "properties/template/add",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr PropertiesTemplateAddAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) PropertiesTemplateAdd(arg *file_properties.AddTemplateArg) (res *file_properties.AddTemplateResult, err error) {
	return dbx.PropertiesTemplateAddContext(context.Background(), arg)
}

// PropertiesTemplateGetAPIError is an error-wrapper for the properties/template/get route
type PropertiesTemplateGetAPIError struct {
	dropbox.APIError
	EndpointError *file_properties.TemplateError `json:"error"`
}

// PropertiesTemplateGetContext : Permission : Team member file access. The
// scope for the route is files.team_metadata.write.
// Deprecated:
func (dbx *apiImpl) PropertiesTemplateGetContext(ctx context.Context, arg *file_properties.GetTemplateArg) (res *file_properties.GetTemplateResult, err error) {
	log.Printf("WARNING: API `PropertiesTemplateGet` is deprecated")

	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "properties/template/get",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr PropertiesTemplateGetAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) PropertiesTemplateGet(arg *file_properties.GetTemplateArg) (res *file_properties.GetTemplateResult, err error) {
	return dbx.PropertiesTemplateGetContext(context.Background(), arg)
}

// ReportsGetActivityAPIError is an error-wrapper for the reports/get_activity route
type ReportsGetActivityAPIError struct {
	dropbox.APIError
	EndpointError *DateRangeError `json:"error"`
}

// ReportsGetActivityContext : Retrieves reporting data about a team's user
// activity. Deprecated: Will be removed on July 1st 2021.
// Deprecated:
func (dbx *apiImpl) ReportsGetActivityContext(ctx context.Context, arg *DateRange) (res *GetActivityReport, err error) {
	log.Printf("WARNING: API `ReportsGetActivity` is deprecated")

	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "reports/get_activity",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr ReportsGetActivityAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) ReportsGetActivity(arg *DateRange) (res *GetActivityReport, err error) {
	return dbx.ReportsGetActivityContext(context.Background(), arg)
}

// ReportsGetDevicesAPIError is an error-wrapper for the reports/get_devices route
type ReportsGetDevicesAPIError struct {
	dropbox.APIError
	EndpointError *DateRangeError `json:"error"`
}

// ReportsGetDevicesContext : Retrieves reporting data about a team's linked
// devices. Deprecated: Will be removed on July 1st 2021.
// Deprecated:
func (dbx *apiImpl) ReportsGetDevicesContext(ctx context.Context, arg *DateRange) (res *GetDevicesReport, err error) {
	log.Printf("WARNING: API `ReportsGetDevices` is deprecated")

	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "reports/get_devices",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr ReportsGetDevicesAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) ReportsGetDevices(arg *DateRange) (res *GetDevicesReport, err error) {
	return dbx.ReportsGetDevicesContext(context.Background(), arg)
}

// ReportsGetMembershipAPIError is an error-wrapper for the reports/get_membership route
type ReportsGetMembershipAPIError struct {
	dropbox.APIError
	EndpointError *DateRangeError `json:"error"`
}

// ReportsGetMembershipContext : Retrieves reporting data about a team's
// membership. Deprecated: Will be removed on July 1st 2021.
// Deprecated:
func (dbx *apiImpl) ReportsGetMembershipContext(ctx context.Context, arg *DateRange) (res *GetMembershipReport, err error) {
	log.Printf("WARNING: API `ReportsGetMembership` is deprecated")

	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "reports/get_membership",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr ReportsGetMembershipAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) ReportsGetMembership(arg *DateRange) (res *GetMembershipReport, err error) {
	return dbx.ReportsGetMembershipContext(context.Background(), arg)
}

// ReportsGetStorageAPIError is an error-wrapper for the reports/get_storage route
type ReportsGetStorageAPIError struct {
	dropbox.APIError
	EndpointError *DateRangeError `json:"error"`
}

// ReportsGetStorageContext : Retrieves reporting data about a team's storage
// usage. Deprecated: Will be removed on July 1st 2021.
// Deprecated:
func (dbx *apiImpl) ReportsGetStorageContext(ctx context.Context, arg *DateRange) (res *GetStorageReport, err error) {
	log.Printf("WARNING: API `ReportsGetStorage` is deprecated")

	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "reports/get_storage",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr ReportsGetStorageAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) ReportsGetStorage(arg *DateRange) (res *GetStorageReport, err error) {
	return dbx.ReportsGetStorageContext(context.Background(), arg)
}

// SharingAllowlistAddAPIError is an error-wrapper for the sharing_allowlist/add route
type SharingAllowlistAddAPIError struct {
	dropbox.APIError
	EndpointError *SharingAllowlistAddError `json:"error"`
}

// SharingAllowlistAddContext : Endpoint adds Approve List entries. Changes are
// effective immediately. Changes are committed in transaction. In case of
// single validation error - all entries are rejected. Valid domains
// (RFC-1034/5) and emails (RFC-5322/822) are accepted. Added entries cannot
// overflow limit of 10000 entries per team. Maximum 100 entries per call is
// allowed.
func (dbx *apiImpl) SharingAllowlistAddContext(ctx context.Context, arg *SharingAllowlistAddArgs) (res *SharingAllowlistAddResponse, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "sharing_allowlist/add",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr SharingAllowlistAddAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) SharingAllowlistAdd(arg *SharingAllowlistAddArgs) (res *SharingAllowlistAddResponse, err error) {
	return dbx.SharingAllowlistAddContext(context.Background(), arg)
}

// SharingAllowlistListAPIError is an error-wrapper for the sharing_allowlist/list route
type SharingAllowlistListAPIError struct {
	dropbox.APIError
	EndpointError *SharingAllowlistListError `json:"error"`
}

// SharingAllowlistListContext : Lists Approve List entries for given team, from
// newest to oldest, returning up to `limit` entries at a time. If there are
// more than `limit` entries associated with the current team, more can be
// fetched by passing the returned `cursor` to `sharingAllowlistListContinue`.
func (dbx *apiImpl) SharingAllowlistListContext(ctx context.Context, arg *SharingAllowlistListArg) (res *SharingAllowlistListResponse, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "sharing_allowlist/list",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr SharingAllowlistListAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) SharingAllowlistList(arg *SharingAllowlistListArg) (res *SharingAllowlistListResponse, err error) {
	return dbx.SharingAllowlistListContext(context.Background(), arg)
}

// SharingAllowlistListContinueAPIError is an error-wrapper for the sharing_allowlist/list/continue route
type SharingAllowlistListContinueAPIError struct {
	dropbox.APIError
	EndpointError *SharingAllowlistListContinueError `json:"error"`
}

// SharingAllowlistListContinueContext : Lists entries associated with given
// team, starting from a the cursor. See `sharingAllowlistList`.
func (dbx *apiImpl) SharingAllowlistListContinueContext(ctx context.Context, arg *SharingAllowlistListContinueArg) (res *SharingAllowlistListResponse, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "sharing_allowlist/list/continue",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr SharingAllowlistListContinueAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) SharingAllowlistListContinue(arg *SharingAllowlistListContinueArg) (res *SharingAllowlistListResponse, err error) {
	return dbx.SharingAllowlistListContinueContext(context.Background(), arg)
}

// SharingAllowlistRemoveAPIError is an error-wrapper for the sharing_allowlist/remove route
type SharingAllowlistRemoveAPIError struct {
	dropbox.APIError
	EndpointError *SharingAllowlistRemoveError `json:"error"`
}

// SharingAllowlistRemoveContext : Endpoint removes Approve List entries.
// Changes are effective immediately. Changes are committed in transaction. In
// case of single validation error - all entries are rejected. Valid domains
// (RFC-1034/5) and emails (RFC-5322/822) are accepted. Entries being removed
// have to be present on the list. Maximum 1000 entries per call is allowed.
func (dbx *apiImpl) SharingAllowlistRemoveContext(ctx context.Context, arg *SharingAllowlistRemoveArgs) (res *SharingAllowlistRemoveResponse, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "sharing_allowlist/remove",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr SharingAllowlistRemoveAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) SharingAllowlistRemove(arg *SharingAllowlistRemoveArgs) (res *SharingAllowlistRemoveResponse, err error) {
	return dbx.SharingAllowlistRemoveContext(context.Background(), arg)
}

// TeamFolderActivateAPIError is an error-wrapper for the team_folder/activate route
type TeamFolderActivateAPIError struct {
	dropbox.APIError
	EndpointError *TeamFolderActivateError `json:"error"`
}

// TeamFolderActivateContext : Sets an archived team folder's status to active.
// Permission : Team member file access.
func (dbx *apiImpl) TeamFolderActivateContext(ctx context.Context, arg *TeamFolderIdArg) (res *TeamFolderMetadata, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "team_folder/activate",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr TeamFolderActivateAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) TeamFolderActivate(arg *TeamFolderIdArg) (res *TeamFolderMetadata, err error) {
	return dbx.TeamFolderActivateContext(context.Background(), arg)
}

// TeamFolderArchiveAPIError is an error-wrapper for the team_folder/archive route
type TeamFolderArchiveAPIError struct {
	dropbox.APIError
	EndpointError *TeamFolderArchiveError `json:"error"`
}

// TeamFolderArchiveContext : Sets an active team folder's status to archived
// and removes all folder and file members. This endpoint cannot be used for
// teams that have a shared team space. This route will either finish
// synchronously, or return a job ID and do the async archive job in background.
// Please use team_folder/archive/check to check the job status. Permission :
// Team member file access.
func (dbx *apiImpl) TeamFolderArchiveContext(ctx context.Context, arg *TeamFolderArchiveArg) (res *TeamFolderArchiveLaunch, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "team_folder/archive",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr TeamFolderArchiveAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) TeamFolderArchive(arg *TeamFolderArchiveArg) (res *TeamFolderArchiveLaunch, err error) {
	return dbx.TeamFolderArchiveContext(context.Background(), arg)
}

// TeamFolderArchiveCheckAPIError is an error-wrapper for the team_folder/archive/check route
type TeamFolderArchiveCheckAPIError struct {
	dropbox.APIError
	EndpointError *async.PollError `json:"error"`
}

// TeamFolderArchiveCheckContext : Returns the status of an asynchronous job for
// archiving a team folder. The job may show '.tag' as complete, but the team
// folder could still be in the process of archiving (indicated by
// `TeamFolderMetadata.status` with 'archive_in_progress'). To confirm that the
// team folder is fully archived, check the field `TeamFolderMetadata.status` in
// the response for the value 'archived'. Permission : Team member file access.
func (dbx *apiImpl) TeamFolderArchiveCheckContext(ctx context.Context, arg *async.PollArg) (res *TeamFolderArchiveJobStatus, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "team_folder/archive/check",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr TeamFolderArchiveCheckAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) TeamFolderArchiveCheck(arg *async.PollArg) (res *TeamFolderArchiveJobStatus, err error) {
	return dbx.TeamFolderArchiveCheckContext(context.Background(), arg)
}

// TeamFolderCreateAPIError is an error-wrapper for the team_folder/create route
type TeamFolderCreateAPIError struct {
	dropbox.APIError
	EndpointError *TeamFolderCreateError `json:"error"`
}

// TeamFolderCreateContext : Creates a new, active, team folder with no members.
// This endpoint can only be used for teams that do not already have a shared
// team space. Permission : Team member file access.
func (dbx *apiImpl) TeamFolderCreateContext(ctx context.Context, arg *TeamFolderCreateArg) (res *TeamFolderMetadata, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "team_folder/create",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr TeamFolderCreateAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) TeamFolderCreate(arg *TeamFolderCreateArg) (res *TeamFolderMetadata, err error) {
	return dbx.TeamFolderCreateContext(context.Background(), arg)
}

// TeamFolderGetInfoAPIError is an error-wrapper for the team_folder/get_info route
type TeamFolderGetInfoAPIError struct {
	dropbox.APIError
	EndpointError struct{} `json:"error"`
}

// TeamFolderGetInfoContext : Retrieves metadata for team folders. Permission :
// Team member file access.
func (dbx *apiImpl) TeamFolderGetInfoContext(ctx context.Context, arg *TeamFolderIdListArg) (res []*TeamFolderGetInfoItem, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "team_folder/get_info",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr TeamFolderGetInfoAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) TeamFolderGetInfo(arg *TeamFolderIdListArg) (res []*TeamFolderGetInfoItem, err error) {
	return dbx.TeamFolderGetInfoContext(context.Background(), arg)
}

// TeamFolderListAPIError is an error-wrapper for the team_folder/list route
type TeamFolderListAPIError struct {
	dropbox.APIError
	EndpointError *TeamFolderListError `json:"error"`
}

// TeamFolderListContext : Lists all team folders. Permission : Team member file
// access.
func (dbx *apiImpl) TeamFolderListContext(ctx context.Context, arg *TeamFolderListArg) (res *TeamFolderListResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "team_folder/list",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr TeamFolderListAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) TeamFolderList(arg *TeamFolderListArg) (res *TeamFolderListResult, err error) {
	return dbx.TeamFolderListContext(context.Background(), arg)
}

// TeamFolderListContinueAPIError is an error-wrapper for the team_folder/list/continue route
type TeamFolderListContinueAPIError struct {
	dropbox.APIError
	EndpointError *TeamFolderListContinueError `json:"error"`
}

// TeamFolderListContinueContext : Once a cursor has been retrieved from
// `teamFolderList`, use this to paginate through all team folders. Permission :
// Team member file access.
func (dbx *apiImpl) TeamFolderListContinueContext(ctx context.Context, arg *TeamFolderListContinueArg) (res *TeamFolderListResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "team_folder/list/continue",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr TeamFolderListContinueAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) TeamFolderListContinue(arg *TeamFolderListContinueArg) (res *TeamFolderListResult, err error) {
	return dbx.TeamFolderListContinueContext(context.Background(), arg)
}

// TeamFolderPermanentlyDeleteAPIError is an error-wrapper for the team_folder/permanently_delete route
type TeamFolderPermanentlyDeleteAPIError struct {
	dropbox.APIError
	EndpointError *TeamFolderPermanentlyDeleteError `json:"error"`
}

// TeamFolderPermanentlyDeleteContext : Permanently deletes an archived team
// folder. This endpoint cannot be used for teams that have a shared team space.
// Permission : Team member file access.
func (dbx *apiImpl) TeamFolderPermanentlyDeleteContext(ctx context.Context, arg *TeamFolderIdArg) (err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "team_folder/permanently_delete",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr TeamFolderPermanentlyDeleteAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	_ = resp
	_ = respBody
	return
}

func (dbx *apiImpl) TeamFolderPermanentlyDelete(arg *TeamFolderIdArg) (err error) {
	return dbx.TeamFolderPermanentlyDeleteContext(context.Background(), arg)
}

// TeamFolderRenameAPIError is an error-wrapper for the team_folder/rename route
type TeamFolderRenameAPIError struct {
	dropbox.APIError
	EndpointError *TeamFolderRenameError `json:"error"`
}

// TeamFolderRenameContext : Changes an active team folder's name. Permission :
// Team member file access.
func (dbx *apiImpl) TeamFolderRenameContext(ctx context.Context, arg *TeamFolderRenameArg) (res *TeamFolderMetadata, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "team_folder/rename",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr TeamFolderRenameAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) TeamFolderRename(arg *TeamFolderRenameArg) (res *TeamFolderMetadata, err error) {
	return dbx.TeamFolderRenameContext(context.Background(), arg)
}

// TeamFolderRestoreAPIError is an error-wrapper for the team_folder/restore route
type TeamFolderRestoreAPIError struct {
	dropbox.APIError
	EndpointError *TeamFolderRestoreError `json:"error"`
}

// TeamFolderRestoreContext : Sets an inactive team folder's status to active.
// Permission: Team member file access.
func (dbx *apiImpl) TeamFolderRestoreContext(ctx context.Context, arg *TeamFolderIdArg) (res *TeamFolderMetadata, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "team_folder/restore",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr TeamFolderRestoreAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) TeamFolderRestore(arg *TeamFolderIdArg) (res *TeamFolderMetadata, err error) {
	return dbx.TeamFolderRestoreContext(context.Background(), arg)
}

// TeamFolderUpdateSyncSettingsAPIError is an error-wrapper for the team_folder/update_sync_settings route
type TeamFolderUpdateSyncSettingsAPIError struct {
	dropbox.APIError
	EndpointError *TeamFolderUpdateSyncSettingsError `json:"error"`
}

// TeamFolderUpdateSyncSettingsContext : Updates the sync settings on a team
// folder or its contents.  Use of this endpoint requires that the team has team
// selective sync enabled.
func (dbx *apiImpl) TeamFolderUpdateSyncSettingsContext(ctx context.Context, arg *TeamFolderUpdateSyncSettingsArg) (res *TeamFolderMetadata, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "team_folder/update_sync_settings",
		Auth:         "team",
		Style:        "rpc",
		Arg:          arg,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr TeamFolderUpdateSyncSettingsAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) TeamFolderUpdateSyncSettings(arg *TeamFolderUpdateSyncSettingsArg) (res *TeamFolderMetadata, err error) {
	return dbx.TeamFolderUpdateSyncSettingsContext(context.Background(), arg)
}

// TokenGetAuthenticatedAdminAPIError is an error-wrapper for the token/get_authenticated_admin route
type TokenGetAuthenticatedAdminAPIError struct {
	dropbox.APIError
	EndpointError *TokenGetAuthenticatedAdminError `json:"error"`
}

// TokenGetAuthenticatedAdminContext : Returns the member profile of the admin
// who generated the team access token used to make the call.
func (dbx *apiImpl) TokenGetAuthenticatedAdminContext(ctx context.Context) (res *TokenGetAuthenticatedAdminResult, err error) {
	req := dropbox.Request{
		Host:         "api",
		Namespace:    "team",
		Route:        "token/get_authenticated_admin",
		Auth:         "team",
		Style:        "rpc",
		Arg:          nil,
		ExtraHeaders: nil,
	}

	var resp []byte
	var respBody io.ReadCloser
	resp, respBody, err = (*dropbox.Context)(dbx).ExecuteContext(ctx, req, nil)
	if err != nil {
		var appErr TokenGetAuthenticatedAdminAPIError
		err = auth.ParseError(err, &appErr)
		if errors.Is(err, &appErr) {
			err = appErr
		}
		return
	}

	err = json.Unmarshal(resp, &res)
	if err != nil {
		return
	}

	_ = respBody
	return
}

func (dbx *apiImpl) TokenGetAuthenticatedAdmin() (res *TokenGetAuthenticatedAdminResult, err error) {
	return dbx.TokenGetAuthenticatedAdminContext(context.Background())
}

// NewContext returns a ContextClient implementation for this namespace
func NewContext(c dropbox.Config) ContextClient {
	ctx := apiImpl(dropbox.NewContext(c))
	return &ctx
}

// New returns a Client implementation for this namespace
func New(c dropbox.Config) Client {
	return NewContext(c)
}
