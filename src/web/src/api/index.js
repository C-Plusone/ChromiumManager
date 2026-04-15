import request from '@/utils/request'

// Group APIs
export const getGroups = () => {
  return request.get(`/get_groups`)
}

export const addGroup = (data) => {
  return request.post(`/add_group`, data)
}

export const updateGroup = (data) => {
  return request.post(`/update_group`, data)
}

export const deleteGroup = (id) => {
  return request.post(`/delete_group`, { id })
}

// Profile APIs
export const getProfiles = (params) => {
  return request.get(`/get_profiles`, { params })
}

export const getProfile = (id) => {
  return request.get(`/get_profile`, { params: { id } })
}

export const addProfile = (data) => {
  return request.post(`/add_profile`, data)
}

export const updateProfile = (data) => {
  return request.post(`/update_profile`, data)
}

export const deleteProfile = (id) => {
  return request.post(`/delete_profile`, { id })
}

export const launchProfile = (data) => {
  return request.post(`/launch_profile`, data)
}

export const stopProfile = (id) => {
  return request.post(`/stop_profile`, { id })
}

export const showProfile = (id) => {
  return request.get(`/show_profile`, { params: { id } })
}

export const exportCookies = (id) => {
  return request.get(`/export_cookies`, { params: { id } })
}

export const importCookies = (data) => {
  return request.post(`/import_cookies`, data)
}

// Template APIs
export const getTemplates = (params) => {
  return request.get(`/get_templates`, { params })
}

export const getTemplate = (id) => {
  return request.get(`/get_template`, { params: { id } })
}

export const addTemplate = (data) => {
  return request.post(`/add_template`, data)
}

export const updateTemplate = (data) => {
  return request.post(`/update_template`, data)
}

export const deleteTemplate = (id) => {
  return request.post(`/delete_template`, { id })
}

export const createFromTemplate = (data) => {
  return request.post(`/create_from_template`, data)
}

export const saveAsTemplate = (data) => {
  return request.post(`/save_as_template`, data)
}

// Proxy APIs
export const getProxies = (params) => {
  return request.get(`/get_proxies`, { params })
}

export const getProxy = (id) => {
  return request.get(`/get_proxy`, { params: { id } })
}

export const addProxy = (data) => {
  return request.post(`/add_proxy`, data)
}

export const updateProxy = (data) => {
  return request.post(`/update_proxy`, data)
}

export const deleteProxy = (id) => {
  return request.post(`/delete_proxy`, { id })
}
