import { request } from './http'

import type {
  CreateSiteUserPayload,
  SiteUser,
  SiteUserListParams,
  SiteUserListResponse,
  UpdateSiteUserPayload,
} from '@/types/site-users'

export function createSiteUser(payload: CreateSiteUserPayload) {
  return request<SiteUser>('/admin/users', { method: 'POST', body: payload })
}

export function listSiteUsers(params: SiteUserListParams = {}) {
  const query = Object.fromEntries(
    Object.entries(params).filter(([, value]) => value !== undefined && value !== ''),
  )
  return request<SiteUserListResponse>('/admin/users', {
    method: 'GET',
    query,
  })
}

export function updateSiteUser(id: number, payload: UpdateSiteUserPayload) {
  return request<SiteUser>(`/admin/users/${id}`, {
    method: 'PUT',
    body: payload,
  })
}
