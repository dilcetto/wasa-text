import { reactive } from 'vue';

const state = reactive({ token: '' });

export function useAuth() {
  function initFromStorage() {
    try {
      const t = localStorage.getItem('token') || '';
      state.token = t;
    } catch {}
  }
  function setToken(t) {
    state.token = t || '';
    try {
      if (state.token) localStorage.setItem('token', state.token);
      else localStorage.removeItem('token');
    } catch {}
  }
  function clearToken() { setToken(''); }
  function getToken() { return state.token; }

  return { initFromStorage, setToken, clearToken, getToken };
}

