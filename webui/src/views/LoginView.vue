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
        const response = await this.$http.post('/login', {
          username: this.username,
        }, {
          headers: {
            'Content-Type': 'application/json',
          },
        });

        const token = response.data.token;

        if (token) {
          localStorage.setItem('token', token);
          localStorage.setItem('username', this.username);
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

