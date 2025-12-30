'use client';

import { createAdminAPIClient, FeatureFlag, Group, UserProfile } from '@/lib/admin/api';
import { getSharedAuthService } from '@/lib/auth/service';
import {
  ChevronLeft,
  ChevronRight,
  Loader2,
  MoreVertical,
  Plus,
  Search,
  Settings,
  Shield,
  Tag,
  Trash2,
  UserCheck,
  Users,
  UserX,
  X,
} from 'lucide-react';
import { useEffect, useState } from 'react';

export default function UserManagementPage() {
  const [users, setUsers] = useState<UserProfile[]>([]);
  const [groups, setGroups] = useState<Group[]>([]);
  const [featureFlags, setFeatureFlags] = useState<FeatureFlag[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [enabledFilter, setEnabledFilter] = useState<boolean | undefined>(undefined);
  const [hideGuestUsers, setHideGuestUsers] = useState(true);
  const [page, setPage] = useState(0);
  const [totalUsers, setTotalUsers] = useState(0);
  const [selectedUser, setSelectedUser] = useState<UserProfile | null>(null);
  const [showGroupModal, setShowGroupModal] = useState(false);
  const [showManageGroupsModal, setShowManageGroupsModal] = useState(false);
  const [showFeatureFlagsModal, setShowFeatureFlagsModal] = useState(false);
  const [selectedGroupForFlags, setSelectedGroupForFlags] = useState<Group | null>(null);
  const [newGroupName, setNewGroupName] = useState('');
  const [selectedGroupId, setSelectedGroupId] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [openMenuUserId, setOpenMenuUserId] = useState<string | null>(null);

  useEffect(() => {
    loadGroups();
    loadFeatureFlags();
  }, []);

  const max = 20;

  useEffect(() => {
    loadGroups();
  }, []);

  useEffect(() => {
    loadUsers();
  }, [page, searchQuery, enabledFilter, hideGuestUsers]);

  async function loadGroups() {
    try {
      const authService = getSharedAuthService();
      const token = await authService.getValidAccessToken();
      if (!token) return;

      const adminClient = createAdminAPIClient(token);
      const response = await adminClient.users.listGroups();
      setGroups(response.data || []);
    } catch (err) {
      console.error('Failed to load groups:', err);
    }
  }

  async function loadFeatureFlags() {
    try {
      const authService = getSharedAuthService();
      const token = await authService.getValidAccessToken();
      if (!token) return;

      const adminClient = createAdminAPIClient(token);
      const response = await adminClient.users.listFeatureFlags();
      setFeatureFlags(response.data || []);
    } catch (err) {
      console.error('Failed to load feature flags:', err);
    }
  }

  // Helper function to find a group by id, name, or path
  // The API returns user's groups as paths like "/guest", "/pilot_users"
  function findGroupByRef(groupRef: string | Group): Group | undefined {
    if (typeof groupRef !== 'string') return groupRef;
    return groups.find((g) => g.id === groupRef || g.name === groupRef || g.path === groupRef);
  }

  // Helper function to check if a user is in a group
  function isUserInGroup(userGroupRef: string | Group, group: Group): boolean {
    if (typeof userGroupRef === 'string') {
      return (
        userGroupRef === group.id || userGroupRef === group.name || userGroupRef === group.path
      );
    }
    return userGroupRef.id === group.id;
  }

  // Helper function to get all feature flags for a user from their groups
  function getUserFeatureFlags(user: UserProfile): string[] {
    if (!user.groups || user.groups.length === 0) return [];

    const flagSet = new Set<string>();
    user.groups.forEach((group) => {
      const groupObj = findGroupByRef(group);

      if (groupObj?.feature_flags) {
        groupObj.feature_flags.forEach((flag) => flagSet.add(flag));
      }
    });

    return Array.from(flagSet);
  }

  async function loadUsers() {
    try {
      setIsLoading(true);
      setError(null);
      const authService = getSharedAuthService();
      const token = await authService.getValidAccessToken();

      if (!token) {
        setError('No authentication token found');
        return;
      }

      const adminClient = createAdminAPIClient(token);
      const response = await adminClient.users.listUsers({
        offset: page * max,
        limit: max,
        search: searchQuery || undefined,
        enabled: enabledFilter,
        exclude_guests: hideGuestUsers,
      });

      setUsers(response.data || []);
      setTotalUsers(response.total || response.data?.length || 0);
    } catch (err) {
      console.error('Failed to load users:', err);
      setError('Failed to load users. Please try again.');
    } finally {
      setIsLoading(false);
    }
  }

  async function handleToggleEnabled(userId: string, currentlyEnabled: boolean) {
    if (
      !confirm(
        `Are you sure you want to ${currentlyEnabled ? 'deactivate' : 'activate'} this user?`,
      )
    ) {
      return;
    }

    try {
      const authService = getSharedAuthService();
      const token = await authService.getValidAccessToken();
      if (!token) return;

      const adminClient = createAdminAPIClient(token);

      if (currentlyEnabled) {
        await adminClient.users.deactivateUser(userId);
      } else {
        await adminClient.users.activateUser(userId);
      }

      loadUsers();
    } catch (err) {
      console.error('Failed to update user status:', err);
      alert('Failed to update user status');
    }
  }

  async function handleAddUserToGroup() {
    if (!selectedUser || !selectedGroupId) return;

    try {
      setIsSubmitting(true);
      const authService = getSharedAuthService();
      const token = await authService.getValidAccessToken();
      if (!token) return;

      const adminClient = createAdminAPIClient(token);
      await adminClient.users.addUserToGroup(selectedUser.id, selectedGroupId);

      // Reload user list and find the updated user
      // Note: getUser() doesn't return groups, but listUsers() does
      const response = await adminClient.users.listUsers({
        offset: 0,
        limit: 100,
        search: selectedUser.email || selectedUser.username,
      });
      const updatedUser = response.data?.find((u) => u.id === selectedUser.id);
      if (updatedUser) {
        setSelectedUser(updatedUser);
      }
      setSelectedGroupId('');
      loadUsers();
    } catch (err) {
      console.error('Failed to add user to group:', err);
      alert('Failed to add user to group');
    } finally {
      setIsSubmitting(false);
    }
  }

  async function handleRemoveUserFromGroup(groupId: string) {
    if (!selectedUser) return;

    try {
      const authService = getSharedAuthService();
      const token = await authService.getValidAccessToken();
      if (!token) return;

      const adminClient = createAdminAPIClient(token);
      await adminClient.users.removeUserFromGroup(selectedUser.id, groupId);

      // Reload user list and find the updated user
      // Note: getUser() doesn't return groups, but listUsers() does
      const response = await adminClient.users.listUsers({
        offset: 0,
        limit: 100,
        search: selectedUser.email || selectedUser.username,
      });
      const updatedUser = response.data?.find((u) => u.id === selectedUser.id);
      if (updatedUser) {
        setSelectedUser(updatedUser);
      }
      loadUsers();
    } catch (err) {
      console.error('Failed to remove user from group:', err);
      alert('Failed to remove user from group');
    }
  }

  async function handleCreateGroup(e: React.FormEvent) {
    e.preventDefault();
    if (!newGroupName.trim()) return;

    try {
      setIsSubmitting(true);
      const authService = getSharedAuthService();
      const token = await authService.getValidAccessToken();
      if (!token) return;

      const adminClient = createAdminAPIClient(token);
      await adminClient.users.createGroup(newGroupName);

      setNewGroupName('');
      loadGroups();
    } catch (err) {
      console.error('Failed to create group:', err);
      alert('Failed to create group');
    } finally {
      setIsSubmitting(false);
    }
  }

  async function handleDeleteGroup(groupId: string) {
    if (!confirm('Are you sure you want to delete this group?')) return;

    try {
      const authService = getSharedAuthService();
      const token = await authService.getValidAccessToken();
      if (!token) return;

      const adminClient = createAdminAPIClient(token);
      await adminClient.users.deleteGroup(groupId);

      loadGroups();
    } catch (err) {
      console.error('Failed to delete group:', err);
      alert('Failed to delete group');
    }
  }

  const totalPages = Math.max(1, Math.ceil(totalUsers / max));
  const userGroups = selectedUser?.groups || [];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h1
            className="text-3xl font-bold tracking-tight flex items-center
gap-2"
          >
            <Users className="w-8 h-8" />
            User Management
          </h1>
          <p className="text-muted-foreground mt-2">Manage users, permissions, and groups</p>
        </div>
        <button
          onClick={() => setShowManageGroupsModal(true)}
          className="flex items-center gap-2 px-4 py-2 border rounded-md
hover:bg-accent transition-colors"
        >
          <Settings className="w-4 h-4" />
          Manage Groups
        </button>
      </div>

      {/* Filters */}
      <div className="bg-card rounded-lg border p-4">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {/* Search */}
          <div className="relative">
            <Search
              className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4
text-muted-foreground"
            />
            <input
              type="text"
              placeholder="Search by email or username..."
              value={searchQuery}
              onChange={(e) => {
                setSearchQuery(e.target.value);
                setPage(0);
              }}
              className="w-full pl-9 pr-3 py-2 border rounded-md bg-background
text-sm focus:outline-none focus:ring-2 focus:ring-primary"
            />
          </div>

          {/* Enabled Filter */}
          <select
            value={enabledFilter === undefined ? 'all' : enabledFilter ? 'enabled' : 'disabled'}
            onChange={(e) => {
              setEnabledFilter(e.target.value === 'all' ? undefined : e.target.value === 'enabled');
              setPage(0);
            }}
            className="px-3 py-2 border rounded-md bg-background text-sm
focus:outline-none focus:ring-2 focus:ring-primary"
          >
            <option value="all">All Users</option>
            <option value="enabled">Enabled Only</option>
            <option value="disabled">Disabled Only</option>
          </select>

          {/* Hide Guest Users */}
          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="hideGuestUsers"
              checked={hideGuestUsers}
              onChange={(e) => {
                setHideGuestUsers(e.target.checked);
                setPage(0);
              }}
              className="rounded border-gray-300 cursor-pointer"
            />
            <label
              htmlFor="hideGuestUsers"
              className="text-sm cursor-pointer
select-none"
            >
              Hide Guest Users
            </label>
          </div>
        </div>

        {/* Clear Filters */}
        {(searchQuery || enabledFilter !== undefined || !hideGuestUsers) && (
          <div className="mt-3">
            <button
              onClick={() => {
                setSearchQuery('');
                setEnabledFilter(undefined);
                setHideGuestUsers(true);
                setPage(0);
              }}
              className="px-4 py-2 text-sm text-muted-foreground
hover:text-foreground border rounded-md hover:bg-accent transition-colors"
            >
              Clear Filters
            </button>
          </div>
        )}
      </div>

      {/* Error */}
      {error && (
        <div
          className="rounded-lg border border-destructive/50
bg-destructive/10 p-4"
        >
          <p className="text-sm text-destructive">{error}</p>
        </div>
      )}

      {/* Users Table */}
      <div className="bg-card rounded-lg border">
        {isLoading ? (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="w-6 h-6 animate-spin text-primary" />
          </div>
        ) : users.length === 0 ? (
          <div className="text-center py-12">
            <Users className="w-12 h-12 mx-auto text-muted-foreground mb-4" />
            <p className="text-muted-foreground">No users found</p>
          </div>
        ) : (
          <div>
            <table className="w-full">
              <thead className="bg-muted/50">
                <tr>
                  <th
                    className="text-left px-4 py-3 text-sm
font-medium"
                  >
                    User
                  </th>
                  <th
                    className="text-left px-4 py-3 text-sm
font-medium"
                  >
                    Email
                  </th>
                  <th
                    className="text-left px-4 py-3 text-sm
font-medium"
                  >
                    Username
                  </th>
                  <th
                    className="text-left px-4 py-3 text-sm
font-medium"
                  >
                    Status
                  </th>
                  <th
                    className="text-left px-4 py-3 text-sm
font-medium"
                  >
                    Groups
                  </th>
                  <th
                    className="text-left px-4 py-3 text-sm
font-medium"
                  >
                    Feature Flags
                  </th>
                  <th
                    className="text-left px-4 py-3 text-sm
font-medium"
                  >
                    Role
                  </th>
                  <th
                    className="text-right px-4 py-3 text-sm
font-medium"
                  >
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y">
                {users.map((user) => {
                  const isEnabled = user.enabled !== false;
                  const displayName =
                    user.first_name && user.last_name
                      ? `${user.first_name} ${user.last_name}`
                      : user.name || user.username || 'Unknown';

                  return (
                    <tr
                      key={user.id}
                      className="hover:bg-muted/30
transition-colors"
                    >
                      <td className="px-4 py-3">
                        <div className="flex items-center gap-3">
                          {user.picture ? (
                            <img
                              src={user.picture}
                              alt={displayName}
                              className="w-8 h-8 rounded-full"
                            />
                          ) : (
                            <div
                              className="w-8 h-8 rounded-full bg-primary/10
flex items-center justify-center"
                            >
                              <Users className="w-4 h-4 text-primary" />
                            </div>
                          )}
                          <span
                            className="font-medium
text-sm"
                          >
                            {displayName}
                          </span>
                        </div>
                      </td>
                      <td className="px-4 py-3 text-sm text-muted-foreground">
                        {user.email || 'N/A'}
                      </td>
                      <td className="px-4 py-3 text-sm text-muted-foreground">
                        {user.username || 'N/A'}
                      </td>
                      <td className="px-4 py-3">
                        <span
                          className={`inline-flex items-center gap-1.5 px-2 py-1 rounded-full text-xs font-medium ${
                            isEnabled
                              ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                              : 'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-400'
                          }`}
                        >
                          <span
                            className={`w-1.5 h-1.5 rounded-full ${isEnabled ? 'bg-green-500' : 'bg-gray-500'}`}
                          />
                          {isEnabled ? 'Enabled' : 'Disabled'}
                        </span>
                      </td>
                      <td className="px-4 py-3">
                        <div className="flex flex-wrap gap-1">
                          {user.groups && user.groups.length > 0 ? (
                            user.groups.slice(0, 2).map((group) => (
                              <span
                                key={typeof group === 'string' ? group : group.id}
                                className="inline-flex items-center gap-1 px-2 py-0.5 rounded-md bg-primary/10 text-primary text-xs"
                              >
                                <Tag className="w-3 h-3" />
                                {typeof group === 'string' ? group : group.name}
                              </span>
                            ))
                          ) : (
                            <span className="text-xs text-muted-foreground">No groups</span>
                          )}
                          {user.groups && user.groups.length > 2 && (
                            <span className="text-xs text-muted-foreground">
                              +{user.groups.length - 2}
                            </span>
                          )}
                        </div>
                      </td>
                      <td className="px-4 py-3">
                        <div className="flex flex-wrap gap-1">
                          {(() => {
                            const userFlags = getUserFeatureFlags(user);
                            return userFlags.length > 0 ? (
                              <>
                                {userFlags.slice(0, 2).map((flag) => (
                                  <span
                                    key={flag}
                                    className="inline-flex items-center gap-1 px-2 py-0.5 rounded-md bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400 text-xs"
                                    title={flag}
                                  >
                                    <Settings className="w-3 h-3" />
                                    {flag.replace(/_/g, ' ')}
                                  </span>
                                ))}
                                {userFlags.length > 2 && (
                                  <span
                                    className="text-xs text-muted-foreground"
                                    title={userFlags.slice(2).join(', ')}
                                  >
                                    +{userFlags.length - 2}
                                  </span>
                                )}
                              </>
                            ) : (
                              <span className="text-xs text-muted-foreground">No flags</span>
                            );
                          })()}
                        </div>
                      </td>
                      <td className="px-4 py-3">
                        {user.is_admin || user.role === 'admin' ? (
                          <span className="inline-flex items-center gap-1 px-2 py-1 rounded-md bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400 text-xs font-medium">
                            <Shield className="w-3 h-3" />
                            Admin
                          </span>
                        ) : (
                          <span className="text-xs text-muted-foreground">User</span>
                        )}
                      </td>
                      <td className="px-4 py-3 relative">
                        <button
                          onClick={(e) => {
                            setOpenMenuUserId(openMenuUserId === user.id ? null : user.id);
                          }}
                          className="p-1 hover:bg-accent rounded transition-colors"
                        >
                          <MoreVertical className="w-4 h-4" />
                        </button>

                        {openMenuUserId === user.id && (
                          <>
                            <div
                              className="fixed inset-0 z-40"
                              onClick={() => setOpenMenuUserId(null)}
                            />
                            <div className="absolute right-0 top-full mt-1 w-48 bg-card border border-border rounded-md shadow-lg z-50">
                              <button
                                onClick={() => {
                                  setSelectedUser(user);
                                  setShowGroupModal(true);
                                  setOpenMenuUserId(null);
                                }}
                                className="w-full flex items-center gap-2 px-4 py-2 text-sm hover:bg-accent transition-colors text-left"
                              >
                                <Tag className="w-4 h-4" />
                                Manage Groups
                              </button>
                              <button
                                onClick={() => {
                                  setSelectedUser(user);
                                  setShowFeatureFlagsModal(true);
                                  setOpenMenuUserId(null);
                                }}
                                className="w-full flex items-center gap-2 px-4 py-2 text-sm hover:bg-accent transition-colors text-left"
                              >
                                <Settings className="w-4 h-4" />
                                View Flags
                              </button>
                              <button
                                onClick={() => {
                                  handleToggleEnabled(user.id, isEnabled);
                                  setOpenMenuUserId(null);
                                }}
                                className={`w-full flex items-center gap-2 px-4 py-2 text-sm hover:bg-accent transition-colors text-left ${
                                  isEnabled ? 'text-destructive' : 'text-green-600'
                                }`}
                              >
                                {isEnabled ? (
                                  <>
                                    <UserX className="w-4 h-4" />
                                    Deactivate
                                  </>
                                ) : (
                                  <>
                                    <UserCheck className="w-4 h-4" />
                                    Activate
                                  </>
                                )}
                              </button>
                            </div>
                          </>
                        )}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between">
          <p className="text-sm text-muted-foreground">
            Showing {page * max + 1} to {Math.min((page + 1) * max, totalUsers)}
            of {totalUsers} users
          </p>
          <div className="flex items-center gap-2">
            <button
              onClick={() => setPage(Math.max(0, page - 1))}
              disabled={page === 0}
              className="p-2 border rounded-md hover:bg-accent
disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              <ChevronLeft className="w-4 h-4" />
            </button>
            <span className="text-sm">
              Page {page + 1} of {totalPages}
            </span>
            <button
              onClick={() => setPage(Math.min(totalPages - 1, page + 1))}
              disabled={page >= totalPages - 1}
              className="p-2 border rounded-md hover:bg-accent
disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              <ChevronRight className="w-4 h-4" />
            </button>
          </div>
        </div>
      )}

      {/* Group Management Modal */}
      {showGroupModal && selectedUser && (
        <div
          className="fixed inset-0 bg-black/50 flex items-center
justify-center z-50 p-4"
        >
          <div className="bg-card rounded-lg border max-w-md w-full p-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-xl font-semibold">Manage Groups</h2>
              <button
                onClick={() => {
                  setShowGroupModal(false);
                  setSelectedUser(null);
                  setSelectedGroupId('');
                }}
                className="p-1 hover:bg-accent rounded-md"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="space-y-4">
              <div>
                <p className="text-sm text-muted-foreground mb-3">
                  User:{' '}
                  <span className="font-medium text-foreground">
                    {selectedUser.username || selectedUser.email}
                  </span>
                </p>

                {/* Current Groups */}
                <div className="space-y-2 mb-4">
                  <label className="block text-sm font-medium">Current Groups</label>
                  {userGroups.length > 0 ? (
                    <div className="space-y-2">
                      {userGroups.map((group) => {
                        // Find the actual group object from groups array
                        // API returns groups as paths like "/guest", "/pilot_users"
                        const groupObj = findGroupByRef(group);

                        if (!groupObj) return null;

                        const groupId = groupObj.id;
                        const groupName = groupObj.name;
                        return (
                          <div
                            key={groupId}
                            className="flex items-center justify-between p-2
border rounded-md"
                          >
                            <span className="text-sm flex items-center gap-2">
                              <Tag className="w-4 h-4 text-primary" />
                              {groupName}
                            </span>
                            <button
                              onClick={() => handleRemoveUserFromGroup(groupId)}
                              className="p-1 hover:bg-destructive/10
text-destructive rounded-md transition-colors"
                            >
                              <Trash2 className="w-4 h-4" />
                            </button>
                          </div>
                        );
                      })}
                    </div>
                  ) : (
                    <p className="text-sm text-muted-foreground py-2">No groups assigned</p>
                  )}
                </div>

                {/* Add Group */}
                <div>
                  <label className="block text-sm font-medium mb-1">Add to Group</label>
                  {(() => {
                    const availableGroups = groups.filter(
                      (g) => !userGroups.some((ug) => isUserInGroup(ug, g)),
                    );

                    if (availableGroups.length === 0) {
                      return (
                        <p className="text-sm text-muted-foreground py-2">
                          {groups.length === 0
                            ? 'No groups available. Create a group first.'
                            : 'User is already in all available groups.'}
                        </p>
                      );
                    }

                    return (
                      <div className="flex gap-2">
                        <select
                          value={selectedGroupId}
                          onChange={(e) => setSelectedGroupId(e.target.value)}
                          className="flex-1 px-3 py-2 border rounded-md
bg-background focus:outline-none focus:ring-2 focus:ring-primary"
                        >
                          <option value="">Select a group...</option>
                          {availableGroups.map((group) => (
                            <option key={group.id} value={group.id}>
                              {group.name}
                            </option>
                          ))}
                        </select>
                        <button
                          onClick={handleAddUserToGroup}
                          disabled={!selectedGroupId || isSubmitting}
                          className="px-3 py-2 bg-primary text-primary-foreground
rounded-md hover:bg-primary/90 transition-colors disabled:opacity-50"
                        >
                          {isSubmitting ? (
                            <Loader2 className="w-4 h-4 animate-spin" />
                          ) : (
                            <Plus className="w-4 h-4" />
                          )}
                        </button>
                      </div>
                    );
                  })()}
                </div>
              </div>

              <div className="flex justify-end pt-4">
                <button
                  onClick={() => {
                    setShowGroupModal(false);
                    setSelectedUser(null);
                    setSelectedGroupId('');
                    loadUsers();
                  }}
                  className="px-4 py-2 bg-primary text-primary-foreground
rounded-md hover:bg-primary/90 transition-colors"
                >
                  Done
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Manage Groups Modal */}
      {showManageGroupsModal && (
        <div
          className="fixed inset-0 bg-black/50 flex items-center
justify-center z-50 p-4"
        >
          <div className="bg-card rounded-lg border max-w-md w-full p-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-xl font-semibold">Manage Groups</h2>
              <button
                onClick={() => setShowManageGroupsModal(false)}
                className="p-1 hover:bg-accent rounded-md"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="space-y-6">
              {/* Create Group Form */}
              <form onSubmit={handleCreateGroup} className="flex gap-2">
                <input
                  type="text"
                  placeholder="New group name..."
                  value={newGroupName}
                  onChange={(e) => setNewGroupName(e.target.value)}
                  className="flex-1 px-3 py-2 border rounded-md bg-background
focus:outline-none focus:ring-2 focus:ring-primary"
                />
                <button
                  type="submit"
                  disabled={!newGroupName.trim() || isSubmitting}
                  className="px-3 py-2 bg-primary text-primary-foreground
rounded-md hover:bg-primary/90 transition-colors disabled:opacity-50"
                >
                  {isSubmitting ? (
                    <Loader2 className="w-4 h-4 animate-spin" />
                  ) : (
                    <Plus className="w-4 h-4" />
                  )}
                </button>
              </form>

              {/* Groups List */}
              <div className="space-y-2">
                <label className="block text-sm font-medium">Existing Groups</label>
                {groups.length > 0 ? (
                  <div className="max-h-60 overflow-y-auto space-y-2 pr-2">
                    {groups.map((group) => (
                      <div
                        key={group.id}
                        className="flex items-center justify-between p-3 border
rounded-md bg-muted/30"
                      >
                        <span className="font-medium flex items-center gap-2">
                          <Tag className="w-4 h-4 text-primary" />
                          {group.name}
                        </span>
                        <button
                          onClick={() => handleDeleteGroup(group.id)}
                          className="p-1.5 hover:bg-destructive/10
text-destructive rounded-md transition-colors"
                          title="Delete group"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      </div>
                    ))}
                  </div>
                ) : (
                  <div
                    className="text-center py-8 border rounded-md
border-dashed"
                  >
                    <Tag className="w-8 h-8 mx-auto text-muted-foreground mb-2" />
                    <p className="text-sm text-muted-foreground">No groups found</p>
                  </div>
                )}
              </div>

              <div className="flex justify-end pt-2">
                <button
                  onClick={() => setShowManageGroupsModal(false)}
                  className="px-4 py-2 border rounded-md hover:bg-accent
transition-colors"
                >
                  Close
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Feature Flags Modal */}
      {showFeatureFlagsModal && selectedUser && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-card rounded-lg border max-w-2xl w-full p-6 max-h-[80vh] overflow-y-auto">
            {/* Header */}
            <div className="flex items-center justify-between mb-4">
              <div>
                <h2 className="text-xl font-bold flex items-center gap-2">
                  <Settings className="w-5 h-5" />
                  Feature Flags
                </h2>
                <p className="text-sm text-muted-foreground mt-1">
                  {selectedUser.first_name} {selectedUser.last_name} ({selectedUser.email})
                </p>
              </div>
              <button
                onClick={() => {
                  setShowFeatureFlagsModal(false);
                  setSelectedUser(null);
                }}
                className="p-1 hover:bg-accent rounded-md transition-colors"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            {/* Content */}
            <div className="space-y-6">
              {/* User's Current Flags */}
              <div>
                <h3 className="font-medium mb-3 flex items-center gap-2">
                  <Shield className="w-4 h-4" />
                  Active Feature Flags
                </h3>

                {(() => {
                  const userFlags = getUserFeatureFlags(selectedUser);
                  return userFlags.length > 0 ? (
                    <div className="space-y-2">
                      {userFlags.map((flag) => {
                        // Find which groups provide this flag
                        const providingGroups = (selectedUser.groups || [])
                          .map((g) =>
                            typeof g === 'string' ? groups.find((gr) => gr.id === g) : g,
                          )
                          .filter((g) => g?.feature_flags?.includes(flag))
                          .map((g) => g!.name);

                        return (
                          <div key={flag} className="p-3 border rounded-md bg-muted/30">
                            <div className="flex items-start justify-between">
                              <div className="flex-1">
                                <div className="font-medium text-sm flex items-center gap-2">
                                  <Settings className="w-4 h-4 text-blue-600" />
                                  {flag}
                                </div>
                                <p className="text-xs text-muted-foreground mt-1">
                                  Provided by: {providingGroups.join(', ')}
                                </p>
                              </div>
                              <span className="inline-flex items-center gap-1 px-2 py-1 rounded-md bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400 text-xs font-medium">
                                Active
                              </span>
                            </div>
                          </div>
                        );
                      })}
                    </div>
                  ) : (
                    <div className="text-center py-8 border rounded-md border-dashed">
                      <Settings className="w-8 h-8 mx-auto text-muted-foreground mb-2" />
                      <p className="text-sm text-muted-foreground">No feature flags active</p>
                      <p className="text-xs text-muted-foreground mt-1">
                        Add this user to groups with feature flags to grant access
                      </p>
                    </div>
                  );
                })()}
              </div>

              {/* User's Groups with Their Flags */}
              <div>
                <h3 className="font-medium mb-3 flex items-center gap-2">
                  <Tag className="w-4 h-4" />
                  Groups & Permissions
                </h3>

                {(() => {
                  const userGroups = selectedUser.groups || [];
                  return userGroups.length > 0 ? (
                    <div className="space-y-3">
                      {userGroups.map((group) => {
                        // API returns groups as paths like "/guest", "/pilot_users"
                        const groupObj = findGroupByRef(group);

                        if (!groupObj) return null;

                        return (
                          <div key={groupObj.id} className="p-3 border rounded-md">
                            <div className="flex items-center justify-between mb-2">
                              <div className="font-medium text-sm flex items-center gap-2">
                                <Tag className="w-4 h-4" />
                                {groupObj.name}
                              </div>
                              <button
                                onClick={() => {
                                  setSelectedGroupForFlags(groupObj);
                                }}
                                className="text-xs text-primary hover:underline"
                              >
                                Manage Flags
                              </button>
                            </div>

                            {groupObj.feature_flags && groupObj.feature_flags.length > 0 ? (
                              <div className="flex flex-wrap gap-1">
                                {groupObj.feature_flags.map((flag) => (
                                  <span
                                    key={flag}
                                    className="inline-flex items-center gap-1 px-2 py-0.5 rounded-md bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400 text-xs"
                                  >
                                    {flag}
                                  </span>
                                ))}
                              </div>
                            ) : (
                              <p className="text-xs text-muted-foreground">
                                No feature flags assigned to this group
                              </p>
                            )}
                          </div>
                        );
                      })}
                    </div>
                  ) : (
                    <div className="text-center py-8 border rounded-md border-dashed">
                      <Tag className="w-8 h-8 mx-auto text-muted-foreground mb-2" />
                      <p className="text-sm text-muted-foreground">User is not in any groups</p>
                    </div>
                  );
                })()}
              </div>

              {/* Available Feature Flags */}
              <div>
                <h3 className="font-medium mb-3 flex items-center gap-2">
                  <Settings className="w-4 h-4" />
                  All Available Feature Flags
                </h3>

                {featureFlags.length > 0 ? (
                  <div className="grid grid-cols-2 gap-2">
                    {featureFlags.map((flag) => {
                      const isActive = getUserFeatureFlags(selectedUser).includes(flag.key);
                      return (
                        <div
                          key={flag.id}
                          className={`p-2 border rounded-md ${
                            isActive
                              ? 'bg-blue-50 dark:bg-blue-950/30 border-blue-200 dark:border-blue-800'
                              : ''
                          }`}
                        >
                          <div className="font-medium text-xs">{flag.name}</div>
                          <div className="text-xs text-muted-foreground">{flag.key}</div>
                          {flag.description && (
                            <div className="text-xs text-muted-foreground mt-1">
                              {flag.description}
                            </div>
                          )}
                        </div>
                      );
                    })}
                  </div>
                ) : (
                  <p className="text-sm text-muted-foreground">No feature flags configured</p>
                )}
              </div>
            </div>

            {/* Footer */}
            <div className="flex justify-between items-center pt-4 mt-4 border-t">
              <p className="text-xs text-muted-foreground">
                💡 Tip: Manage feature flags through groups for better organization
              </p>
              <button
                onClick={() => {
                  setShowFeatureFlagsModal(false);
                  setSelectedUser(null);
                }}
                className="px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Group Feature Flags Management Modal */}
      {selectedGroupForFlags && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-60 p-4">
          <div className="bg-card rounded-lg border max-w-xl w-full p-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-bold flex items-center gap-2">
                <Tag className="w-5 h-5" />
                Manage Feature Flags - {selectedGroupForFlags.name}
              </h3>
              <button
                onClick={() => setSelectedGroupForFlags(null)}
                className="p-1 hover:bg-accent rounded-md transition-colors"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="space-y-4">
              <p className="text-sm text-muted-foreground">
                Select feature flags to enable for all members of this group.
              </p>

              {/* Feature Flag Checkboxes */}
              <div className="space-y-2 max-h-96 overflow-y-auto">
                {featureFlags.map((flag) => {
                  const isEnabled =
                    selectedGroupForFlags.feature_flags?.includes(flag.key) || false;

                  return (
                    <label
                      key={flag.id}
                      className="flex items-start gap-3 p-3 border rounded-md hover:bg-muted/50 cursor-pointer transition-colors"
                    >
                      <input
                        type="checkbox"
                        checked={isEnabled}
                        onChange={async (e) => {
                          try {
                            const authService = getSharedAuthService();
                            const token = await authService.getValidAccessToken();
                            if (!token) return;

                            const adminClient = createAdminAPIClient(token);

                            if (e.target.checked) {
                              // Enable flag
                              await adminClient.users.setGroupFeatureFlags(
                                selectedGroupForFlags.id,
                                [...(selectedGroupForFlags.feature_flags || []), flag.key],
                              );
                            } else {
                              // Disable flag
                              await adminClient.users.setGroupFeatureFlags(
                                selectedGroupForFlags.id,
                                (selectedGroupForFlags.feature_flags || []).filter(
                                  (f) => f !== flag.key,
                                ),
                              );
                            }

                            // Reload groups and users
                            await loadGroups();
                            await loadUsers();

                            // Update selected group
                            const updatedGroup = groups.find(
                              (g) => g.id === selectedGroupForFlags.id,
                            );
                            if (updatedGroup) {
                              setSelectedGroupForFlags(updatedGroup);
                            }
                          } catch (err) {
                            console.error('Failed to update feature flag:', err);
                            alert('Failed to update feature flag');
                          }
                        }}
                        className="mt-0.5 rounded border-gray-300"
                      />
                      <div className="flex-1">
                        <div className="font-medium text-sm">{flag.name}</div>
                        <div className="text-xs text-muted-foreground">{flag.key}</div>
                        {flag.description && (
                          <div className="text-xs text-muted-foreground mt-1">
                            {flag.description}
                          </div>
                        )}
                      </div>
                    </label>
                  );
                })}
              </div>

              <div className="flex justify-end pt-2 border-t">
                <button
                  onClick={() => setSelectedGroupForFlags(null)}
                  className="px-4 py-2 border rounded-md hover:bg-accent transition-colors"
                >
                  Done
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
