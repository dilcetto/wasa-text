<template>
  <div class="login-container d-flex justify-content-center align-items-center">
    <div class="card p-4 login card">
      <div class="mb-3">
        <input 
          v-model="username"
          type="text"
          class="form-control"
          placeholder="Enter your username"
        />
      </div>
      <ErrorMsg v-if="error" :msg="error" />
      <button class="btn btn-primary w-100" @click="doLogin">Login</button>
    </div>
  </div>
</template>

<script>
export default {
  data() {
    localStorage.clear();
    return {
      error: null,
      username: '',
    };
  },
  methods: {
    async doLogin() {
      if (this.username.trim() === '') {
        this.error = "Username cannot be empty";
        return;
      }

      try {
        const response = await this.$axios.post('/login', {
          username: this.username,
        }, {
          headers: {
            'Content-Type': 'application/json',
          },
        });

        const {token, user} = response.data;

        if (token && user) {
          localStorage.setItem('token', token);
          localStorage.setItem('username', user.username);
          localStorage.setItem('userId', user.id);
          this.$router.push('/home');
        } else {
          this.error = 'Login failed: No token received';
        }

      } catch (error) {
        if (error.response && error.response.status === 400) {
          this.error = 'Login failed: Invalid username';
        } else {
          this.error = 'Login failed: ' + (error.message || 'Unknown error');
        }
      }
    },
  },
};
</script>

<style scoped>
.login-container {
  height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  background: var(--bg); /* dark background */
  color: var(--text);
  font-family: 'Segoe UI', Roboto, sans-serif;
}

.card {
  max-width: 400px;
  width: 100%;
  background: var(--bg-alt); /* slightly lighter than main bg */
  border: 1px solid var(--border);
  border-radius: var(--radius);
  box-shadow: var(--shadow);
}

.form-control {
  background: var(--bg-hover);
  border: 1px solid var(--border);
  color: var(--text);
  border-radius: var(--radius);
  padding: 0.75rem;
  transition: 0.2s;
}
.form-control:focus {
  background: #2d2d2d;
  border-color: var(--accent);
  box-shadow: 0 0 8px var(--accent);
  color: var(--text);
}

.btn-primary {
  background: var(--accent);
  border: none;
  border-radius: var(--radius);
  padding: 0.75rem;
  font-weight: bold;
  transition: 0.2s;
}
.btn-primary:hover {
  background: var(--accent-alt); /* switch to neon cyan on hover */
  box-shadow: 0 0 10px var(--accent-alt);
}
</style>