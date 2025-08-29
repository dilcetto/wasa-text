<template>
  <section class="search">
    <header class="search-header">
      <h2>Search</h2>
    </header>

    <div class="search-form">
      <input v-model="qUser" class="input" type="text" placeholder="Search user (username)" />
      <input v-model="qConv" class="input" type="text" placeholder="Search conversation (name)" />
      <button class="btn" :disabled="!canSearch || loading" @click="doSearch">{{ loading ? 'Searching…' : 'Search' }}</button>
    </div>

    <ErrorMsg v-if="error" :msg="error" />

    <div class="results">
      <div class="col">
        <h3>Users</h3>
        <div v-if="users.length === 0" class="muted">No users</div>
        <div v-for="u in users" :key="u.id" class="row">
          <div class="avatar" v-if="u.photo"><img :src="'data:image/png;base64,' + u.photo" alt="" /></div>
          <div class="label">@{{ u.username }}</div>
        </div>
      </div>
      <div class="col">
        <h3>Conversations</h3>
        <div v-if="conversations.length === 0" class="muted">No conversations</div>
        <div v-for="c in conversations" :key="c.conversationId" class="row link" @click="openConv(c.conversationId)">
          <div class="avatar" v-if="c.profilePhoto"><img :src="'data:image/png;base64,' + c.profilePhoto" alt="" /></div>
          <div class="label">{{ c.displayName }}</div>
        </div>
      </div>
    </div>
  </section>
</template>

<script>
export default {
  name: 'SearchView',
  data() {
    return {
      qUser: '',
      qConv: '',
      users: [],
      conversations: [],
      error: null,
      loading: false,
    };
  },
  computed: {
    canSearch() {
      return (this.qUser && this.qUser.trim().length > 0) || (this.qConv && this.qConv.trim().length > 0);
    },
  },
  methods: {
    async doSearch() {
      if (!this.canSearch) return;
      this.loading = true;
      this.error = null;
      this.users = [];
      this.conversations = [];
      try {
        const token = localStorage.getItem('token');
        if (!token) {
          this.$router.push('/login');
          return;
        }
        const params = new URLSearchParams();
        if (this.qUser.trim()) params.set('user', this.qUser.trim());
        if (this.qConv.trim()) params.set('conversation', this.qConv.trim());
        const res = await this.$axios.get(`/searchby?${params.toString()}`, {
          headers: { Authorization: `Bearer ${token}` },
        });
        this.users = res.data?.users || [];
        this.conversations = res.data?.conversations || [];
      } catch (e) {
        console.error('Search failed', e);
        this.error = 'Search failed';
      } finally {
        this.loading = false;
      }
    },
    openConv(id) {
      if (!id) return;
      this.$router.push(`/conversations/${id}`);
    },
  },
};
</script>

<style scoped>
.search { display: grid; gap: 1rem; }
.search-header { border-bottom: 1px solid var(--border); padding-bottom: .5rem; }
.search-form { display: flex; gap: .5rem; align-items: center; }
.input { flex: 1; padding: .6rem .7rem; border: 1px solid var(--border); border-radius: var(--radius); background: var(--bg); color: var(--text); }
.btn { padding: .55rem .9rem; border-radius: var(--radius); border: none; background: var(--accent); color: #000; font-weight: 700; }
.results { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
.col { background: var(--bg-alt); border: 1px solid var(--border); border-radius: var(--radius); padding: .75rem; }
.row { display: flex; align-items: center; gap: .5rem; padding: .4rem .2rem; }
.row.link { cursor: pointer; }
.row.link:hover { background: var(--bg-hover); border-radius: .5rem; }
.avatar { width: 32px; height: 32px; border-radius: 50%; overflow: hidden; background: var(--bg); display: grid; place-items: center; }
.avatar img { width: 100%; height: 100%; object-fit: cover; }
.label { color: var(--text); }
.muted { color: var(--text-dim); }
</style>

