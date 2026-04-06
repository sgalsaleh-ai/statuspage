const API_BASE = '/api';

function getHeaders() {
  const headers = { 'Content-Type': 'application/json' };
  const token = localStorage.getItem('token');
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  return headers;
}

async function request(path, options = {}) {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: getHeaders(),
    ...options,
  });
  if (res.status === 401) {
    localStorage.removeItem('token');
    window.location.href = '/admin/login';
    throw new Error('Unauthorized');
  }
  const data = await res.json();
  if (!res.ok) {
    throw new Error(data.error || 'Request failed');
  }
  return data;
}

export const api = {
  // Auth
  checkAuth: () => request('/auth/check'),
  login: (username, password) =>
    request('/auth/login', { method: 'POST', body: JSON.stringify({ username, password }) }),
  setup: (password) =>
    request('/auth/setup', { method: 'POST', body: JSON.stringify({ password }) }),

  // Public
  getStatus: () => request('/status'),
  getIncidents: () => request('/incidents'),
  getIncident: (id) => request(`/incidents/${id}`),
  subscribe: (email) =>
    request('/subscribers', { method: 'POST', body: JSON.stringify({ email }) }),

  // Admin - Components
  getComponents: () => request('/admin/components'),
  createComponent: (data) =>
    request('/admin/components', { method: 'POST', body: JSON.stringify(data) }),
  updateComponent: (id, data) =>
    request(`/admin/components/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteComponent: (id) =>
    request(`/admin/components/${id}`, { method: 'DELETE' }),

  // Admin - Incidents
  getAdminIncidents: () => request('/admin/incidents'),
  createIncident: (data) =>
    request('/admin/incidents', { method: 'POST', body: JSON.stringify(data) }),
  updateIncident: (id, data) =>
    request(`/admin/incidents/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteIncident: (id) =>
    request(`/admin/incidents/${id}`, { method: 'DELETE' }),
  createIncidentUpdate: (id, data) =>
    request(`/admin/incidents/${id}/updates`, { method: 'POST', body: JSON.stringify(data) }),

  // Admin - Subscribers
  getSubscribers: () => request('/admin/subscribers'),
  deleteSubscriber: (id) =>
    request(`/admin/subscribers/${id}`, { method: 'DELETE' }),

  // Centrifugo
  getCentrifugoConfig: () => request('/centrifugo/config'),
};
