<template>
  <section class="group-edit">
    <header class="top">
      <h1 class="title">Edit Group</h1>
    </header>

    <div class="card">
      <ErrorMsg v-if="errormsg" :msg="errormsg" />

      <div class="field">
        <label>Members</label>
        <ul class="members">
          <li v-for="m in visibleMembers" :key="m.id">
            <span class="pill">@{{ m.username }}</span>
          </li>
        </ul>
      </div>

      <div class="field">
        <label>Current Name</label>
        <div class="dim">{{ group.displayName || '—' }}</div>
      </div>

      <div class="field">
        <label>New Name</label>
        <input v-model="newName" class="input" placeholder="Enter new group name" />
        <div class="actions">
          <button class="btn btn-primary" :disabled="!canRename || renaming" @click="updateName">
            {{ renaming ? 'Updating…' : 'Update Name' }}
          </button>
        </div>
      </div>

      <div class="field">
        <label>Group Photo</label>
        <div class="photo-row">
          <img v-if="currentPhoto" :src="currentPhoto" alt="Group" class="avatar" />
          <div class="upload">
            <input type="file" accept="image/*" @change="handlePhoto" />
            <button class="btn" :disabled="!newPhoto || updatingPhoto" @click="updatePhoto">
              {{ updatingPhoto ? 'Updating…' : 'Update Photo' }}
            </button>
          </div>
        </div>
      </div>

      <div class="field">
        <label>Add Member</label>
        <div class="row">
          <input v-model="memberUsername" class="input" placeholder="Username to add" />
          <button class="btn" :disabled="!canAdd || adding" @click="addMember">{{ adding ? 'Adding…' : 'Add' }}</button>
        </div>
      </div>

      <div class="field">
        <button class="btn btn-danger" @click="leaveGroup">Leave Group</button>
      </div>

      <div v-if="success" class="success">Saved successfully.</div>
    </div>
  </section>
</template>

<script>
import ErrorMsg from '@/components/ErrorMsg.vue';

export default {
  name: 'GroupEditView',
  components: { ErrorMsg },
  data() {
    return {
      group: {},
      members: [],
      newName: '',
      newPhoto: '',
      memberUsername: '',
      renaming: false,
      updatingPhoto: false,
      adding: false,
      errormsg: null,
      success: false,
    };
  },
  computed: {
    groupId() {
      return this.$route.params.groupId;
    },
    currentUserId() {
      return localStorage.getItem('userId') || '';
    },
    visibleMembers() {
      try { return (this.members || []).filter(m => m.id !== this.currentUserId); } catch { return []; }
    },
    // members fetched separately
    currentPhoto() {
      const b64 = this.group?.profilePhoto || '';
      return b64 ? 'data:image/png;base64,' + b64 : 'nopfp.jpg';
    },
    canRename() {
      const v = this.newName?.trim();
      return !!v && v.length >= 3 && v.length <= 50;
    },
    canAdd() {
      const v = (this.memberUsername || '').trim();
      const myName = localStorage.getItem('username') || '';
      return !!v && v.length >= 3 && v.length <= 16 && v.toLowerCase() !== myName.toLowerCase();
    },
  },
  methods: {
    async load() {
      this.errormsg = null;
      try {
        const res = await this.$axios.get(`/conversations/${this.groupId}`);
        this.group = res.data || {};
        await this.loadMembers();
      } catch (e) {
        this.errormsg = 'Failed to load group';
      }
    },
    async loadMembers() {
      try {
        const res = await this.$axios.get(`/conversations/${this.groupId}/members`);
        this.members = res.data || [];
      } catch (e) {
        // non-blocking
      }
    },
    async updateName() {
      if (!this.canRename || this.renaming) return;
      this.renaming = true;
      this.errormsg = null;
      this.success = false;
      try {
        await this.$axios.put(`/groups/${this.groupId}/name`, { newName: this.newName.trim() });
        this.success = true;
        this.newName = '';
        await this.load();
      } catch (e) {
        this.errormsg = 'Failed to update name';
      } finally {
        this.renaming = false;
      }
    },
    handlePhoto(e) {
      const file = e?.target?.files?.[0];
      if (!file) { this.newPhoto = ''; return; }
      if (!file.type.startsWith('image/')) {
        this.errormsg = 'Please select an image file.';
        return;
      }
      const max = 10 * 1024 * 1024;
      if (file.size > max) { this.errormsg = 'Image too large (max 10MB)'; if (e?.target) e.target.value=''; return; }
      const reader = new FileReader();
      reader.onload = () => {
        const result = reader.result || '';
        this.newPhoto = typeof result === 'string' && result.includes(',') ? result.split(',')[1] : result;
      };
      reader.onerror = () => { this.errormsg = 'Error reading file.' };
      reader.readAsDataURL(file);
    },
    async updatePhoto() {
      if (!this.newPhoto || this.updatingPhoto) return;
      this.updatingPhoto = true;
      this.errormsg = null;
      this.success = false;
      try {
        await this.$axios.put(`/groups/${this.groupId}/photo`, { groupPhoto: this.newPhoto });
        this.success = true;
        this.newPhoto = '';
        await this.load();
      } catch (e) {
        this.errormsg = 'Failed to update photo';
      } finally {
        this.updatingPhoto = false;
      }
    },
    async addMember() {
      if (!this.canAdd || this.adding) return;
      this.adding = true;
      this.errormsg = null;
      this.success = false;
      try {
        await this.$axios.post(`/groups/${this.groupId}`, { username: this.memberUsername.trim() });
        this.success = true;
        this.memberUsername = '';
        await this.load();
      } catch (e) {
        this.errormsg = 'Failed to add member';
      } finally {
        this.adding = false;
      }
    },
    async leaveGroup() {
      if (!confirm('Leave this group?')) return;
      this.errormsg = null;
      try {
        const userId = localStorage.getItem('userId');
        await this.$axios.delete(`/groups/${this.groupId}`, { data: { user_id: userId } });
        this.$router.push('/home');
      } catch (e) {
        this.errormsg = 'Failed to leave group';
      }
    },
  },
  mounted() {
    this.load();
  },
};
</script>

<style scoped>

.group-edit {
  min-height: 100vh;
  background: var(--bg);
  color: var(--text);
  padding: 1rem;
  display: grid;
  gap: 1rem;
}


.top {
  border-bottom: 1px solid var(--border);
  padding-bottom: .5rem;
}
.title {
  margin: 0;
  font-weight: 800;
  font-size: clamp(1.1rem, 1rem + .6vw, 1.5rem);
}


.card {
  max-width: 760px;
  background: var(--bg-alt);
  border: 1px solid var(--border);
  border-radius: 14px;
  padding: 1rem;
  box-shadow: 0 10px 30px rgba(0,0,0,.35);
  display: grid;
  gap: 1rem;
}

/* Fields */
.field { display: grid; gap: .35rem; }
label { font-weight: 600; }
.dim { color: var(--text-dim); }

/* Inputs */
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
.input.invalid {
  border-color: #ef4444;
  box-shadow: 0 0 8px rgba(239,68,68,.35);
}

.actions {
  margin-top: .4rem;
  display: flex;
  gap: .6rem;
}


.photo-row {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 1rem;
  align-items: center;
}
.avatar {
  width: 80px; height: 80px;
  border-radius: 10px;
  object-fit: cover;
  background: var(--bg);
  border: 2px solid var(--accent-alt);
  box-shadow: 0 0 10px var(--accent-alt);
}
.upload {
  display: flex;
  gap: .6rem;
  align-items: center;
  flex-wrap: wrap;
}
.upload input[type="file"] {
  color: var(--text-dim);
  max-width: 320px;
}


.btn {
  appearance: none;
  cursor: pointer;
  background: color-mix(in oklab, var(--bg) 85%, var(--accent) 15%);
  color: var(--text);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: .55rem .85rem;
  transition: background .15s ease, transform .06s ease, box-shadow .15s ease;
}
.btn:hover { background: var(--bg-hover); transform: translateY(-1px); }
.btn:active { transform: translateY(0); }
.btn[disabled] { opacity: .6; cursor: not-allowed; }

.btn-primary {
  background: linear-gradient(90deg, var(--accent), var(--accent-alt));
  color: #0b0f17;
  border: 1px solid color-mix(in oklab, var(--accent) 60%, var(--accent-alt) 40%);
  box-shadow: 0 0 10px var(--accent), 0 0 16px var(--accent-alt);
}
.btn-primary:hover { filter: saturate(1.05) brightness(1.03); }

.success {
  margin-top: .25rem;
  padding: .6rem .8rem;
  border-radius: var(--radius);
  border: 1px solid color-mix(in oklab, #16a34a 60%, var(--border) 40%);
  background: color-mix(in oklab, var(--bg) 80%, #16a34a 20%);
  color: #eaffea;
  box-shadow: 0 0 10px rgba(22,163,74,.25);
  font-weight: 600;
}

.members { list-style: none; padding-left: 0; display: flex; gap: .5rem; flex-wrap: wrap; }
.pill { background: var(--bg-hover); border: 1px solid var(--border); border-radius: 999px; padding: .2rem .6rem; }
.you { color: var(--text-dim); margin-left: .25rem; }


@media (max-width: 640px) {
  .photo-row { grid-template-columns: 1fr; justify-items: start; }
  .avatar { width: 72px; height: 72px; border-radius: 8px; }
}
</style>
