import axios from "axios";

const instance = axios.create({
    baseURL: __API_URL__,
    timeout: 1000 * 7
});

// Set Authorization from localStorage at startup (single source: 'token')
try {
  const t = localStorage.getItem("token");
  if (t) instance.defaults.headers.common["Authorization"] = `Bearer ${t}`;
} catch {}

export default instance;
