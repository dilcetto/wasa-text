<template>
  <section class="group-create">
    <header class="top">
      <h1 class="title">Create Group</h1>
    </header>

    <div class="card">
      <ErrorMsg v-if="errormsg" :msg="errormsg" />

      <div class="form-group">
        <label for="groupName">Group Name</label>
        <input
          v-model="groupName"
          id="groupName"
          type="text"
          class="input-field"
          placeholder="Enter group name"
        />
      </div>

      <div class="form-group">
        <label>Group Photo (optional)</label>
        <input type="file" @change="handlePhoto" accept="image/*" />
      </div>
      <div class="form-group">
        <label>Add Members</label>
        <input
          v-model="userSearch"
          class="input"
          placeholder="Search users..."
          @input="searchUsers"
        />
        <ul v-if="searchResults.length">
          <li v-for="user in searchResults" :key="user.id">
            {{ user.username }}
            <button @click="addMember(user)">Add</button>
          </li>
        </ul>
        <div>
          <span v-for="member in members" :key="member.id" class="member-chip">
            {{ member.username }}
            <button @click="removeMember(member)">x</button>
          </span>
        </div>
      </div>
      <button
        class="btn btn-primary"
        @click="createGroup"
        :disabled="!canCreate"
      >
        Create Group
      </button>
      <div v-if="success" class="success-msg">Group created!</div>
    </div>
  </section>
</template>

<script>
import ErrorMsg from '@/components/ErrorMsg.vue';

export default {
  name: "GroupCreateView",
  components: { ErrorMsg },
  data() {
    return {
      groupName: "",
      groupPhoto: "",
      userSearch: "",
      searchResults: [],
      members: [],
      errormsg: null,
      success: false,
      currentUserId: localStorage.getItem('userId') || '',
    };
  },
  computed: {
    canCreate() {
      return this.groupName.trim().length >= 3 && this.members.length > 0;
    }
  },
  methods: {
    handlePhoto(e) {
      const file = e?.target?.files?.[0];
      if (!file) { this.groupPhoto = ""; return; }
      if (!file.type.startsWith('image/')) { this.errormsg = 'Please select an image file'; return; }
      const max = 10 * 1024 * 1024;
      if (file.size > max) { this.errormsg = 'Image too large (max 10MB)'; if (e?.target) e.target.value=''; return; }
      const reader = new FileReader();
      reader.onload = () => {
        const result = reader.result || '';
        // Store base64 only (no data URL prefix)
        const b64 = typeof result === 'string' && result.includes(',') ? result.split(',')[1] : result;
        this.groupPhoto = b64;
      };
      reader.onerror = () => { this.errormsg = 'Failed to read file'; };
      reader.readAsDataURL(file);
    },
    async searchUsers() {
      if (!this.userSearch.trim()) {
        this.searchResults = [];
        return;
      }
      try {
        const token = localStorage.getItem('token');
        const res = await this.$axios.get(`/searchby?user=${encodeURIComponent(this.userSearch)}`, {
          headers: { Authorization: `Bearer ${token}` }
        });
        const all = res.data?.users || [];
        // exclude the current user and already-added members from search results
        this.searchResults = all.filter(u => u.id !== this.currentUserId && !this.members.find(m => m.id === u.id));
      } catch (e) {
        this.errormsg = "Failed to search users";
      }
    },
    addMember(user) {
      // backend includes creator automatically
      if (user?.id === this.currentUserId) return;
      if (!this.members.find(m => m.id === user.id)) {
        this.members.push(user);
      }
    },
    removeMember(user) {
      this.members = this.members.filter(m => m.id !== user.id);
    },
    async createGroup() {
      this.errormsg = null;
      this.success = false;
      try {
        const token = localStorage.getItem('token');
        const payload = {
          groupName: this.groupName,
          // send other members 
          members: this.members.map(m => m.id).filter(id => id && id !== this.currentUserId),
          groupPhoto: this.groupPhoto || undefined,
        };
        const res = await this.$axios.post('/groups', payload, {
          headers: { Authorization: `Bearer ${token}` }
        });
        this.success = true;
        this.groupName = "";
        this.groupPhoto = "";
        this.members = [];
        const convId = res?.data?.conversationId;
        if (convId) this.$router.push(`/conversations/${convId}`);
      } catch (e) {
        this.errormsg = "Failed to create group";
      }
    }
  }
};
</script>

<style scoped>
.group-create {
    min-height: 100vh;
    background: var(--bg);
    color: var(--text);
    padding: 1rem;
    display: grid;
    gap: 1rem;
}

.top { border-bottom: 1px solid var(--border); padding-bottom: .5rem; }
.title { margin: 0; font-size: clamp(1.1rem, 1rem + .6vw, 1.5rem); font-weight: 800; }

.card {
  max-width: 720px;
  background: var(--bg-alt);
  border: 1px solid var(--border);
  border-radius: 14px;
  padding: 1rem;
  box-shadow: 0 10px 30px rgba(0,0,0,.35);
}

.form { display: grid; gap: 1rem; }
.field { display: grid; gap: .35rem; }
label { font-weight: 600; }
.dim { color: var(--text-dim); }

.input {
  background: var(--bg-hover);
  border: 1px solid var(--border);
  color: var(--text);
  border-radius: var(--radius);
  padding: .6rem .7rem;
  transition: .15s ease;
}
.input::placeholder { color: var(--text-dim); }
.input:focus {
  outline: none;
  border-color: var(--accent);
  box-shadow: 0 0 8px var(--accent);
  background: #2d2d2d;
}
.input.invalid { border-color: #ef4444; box-shadow: 0 0 8px rgba(239,68,68,.35); }
.textarea { resize: vertical; min-height: 96px; }

.hint { font-size: .85rem; color: var(--text-dim); }
.hint.bad { color: #ffb0b0; }

.actions {
  display: flex; justify-content: flex-end; gap: .6rem;
}
.btn {
  appearance: none; cursor: pointer;
  background: color-mix(in oklab, var(--bg) 85%, var(--accent) 15%);
  color: var(--text);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: .6rem .9rem;
  transition: background .15s ease, transform .06s ease, box-shadow .15s ease;
}
.btn:hover { background: var(--bg-hover); transform: translateY(-1px); }
.btn:active { transform: translateY(0); }
.btn[disabled] { opacity: .6; cursor: not-allowed; }
.btn.btn-primary {
  background: linear-gradient(90deg, var(--accent), var(--accent-alt));
  color: #0b0f17;
  border: 1px solid color-mix(in oklab, var(--accent) 60%, var(--accent-alt) 40%);
  box-shadow: 0 0 10px var(--accent), 0 0 16px var(--accent-alt);
}
</style>
