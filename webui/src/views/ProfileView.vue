<template>
    <section class="profile">
        <!--topbar-->
        <header class="profile-header">
            <h1 class="title">
                <span class="hello">{{ username }}</span>
                <span class="muted">, Welcome to your profile</span>
            </h1>
            <div class="profile-actions">
                <button class="btn" @click="editProfile = !editProfile">
                    {{ editProfile ? 'Cancel' : 'Edit Profile' }}
                </button>
                <button class="btn" @click="startChat">Start Chat</button>
                <button class="btn btn-danger" @click="logOut">Log Out</button>
            </div>
        </header>

        <!--container-->
        <div class="profile-container">
            <div class="profile-head">
                <div class="profile-photo">
                    <img 
                        :src="user.photo || 'nopfp.jpg'"
                        alt="Profile Photo" 
                        class="rounded-circle"
                        width="100"
                        height="100"
                    />
                </div>
                <div class="username-container">
                    <h2 class="username">{{ username }}</h2>
                     <div v-if="editProfile" class="edit-profile">
                        <div class="update-username">
                            <input
                                v-model="newUsername"
                                class="input"
                                placeholder="Enter new username" 
                                maxlength="16" 
                                minlength="3"
                            />
                            <button class="btn btn-primary" @click="updateUsername" :disabled="!canUpdateUsername">Update Username</button>
                        </div>
                 <div class="update-photo">
                    <input type="file" accept="image/*" @change="handlePhotoUpload" />
                    <button class="btn btn-primary" @click="updatePhoto" :disabled="!newPhoto">Update Photo</button>
                </div>
            </div>
            <ErrorMsg v-if="errormsg" :msg="errormsg" />
        </div>
      </div>
    </div>
  </section>
</template>

<script>
import axios from "../services/axios";
import ErrorMsg from '@/components/ErrorMsg.vue';

export default {
    name: 'ProfileView',
    components: { ErrorMsg },
    data() {
        return {
            editProfile: false,
            errormsg: null,

            //user state
            user: {
                photo: localStorage.getItem('userPhoto') || 'nopfp.jpg',
            },
            username: localStorage.getItem('username') || '',
            newUsername: '',
            newPhoto: null,
        }
    },
    computed: {
        canUpdateUsername() {
            const v = this.newUsername.trim();
            return !! v && v !== this.username && v.length >= 3 && v.length <= 16;
        },
    },
    methods: {
        initFromLocal() {
          const name = localStorage.getItem('username') || '';
          const photo = localStorage.getItem('userPhoto') || 'nopfp.jpg';
          this.username = name;
          this.user.photo = photo;
        },
        startChat() {
            this.$router.push('/search');
        },
        logOut() {
            localStorage.clear();
            try { delete this.$axios.defaults.headers.common['Authorization'] } catch (e) {}
            this.$router.push('/login');
        },
 
        async updateUsername() {
            if (!this.newUsername || this.newUsername === this.username) return;
            try {
              await axios.put('/user/username', { username: this.newUsername });
                alert('Username updated successfully!');
                this.username = this.newUsername;
                localStorage.setItem('username', this.username);
                this.newUsername = '';
                this.errormsg = null;
            } catch (error) {
                console.error('Error updating username:', error);
                const data = error?.response?.data;
                this.errormsg = (typeof data === 'object' ? (data?.error || data?.message) : data) || 'Failed to update username.';
            }
        },
        refresh() {
            this.fetchProfile();
        },

        handlePhotoUpload(e) {
            const file = e?.target?.files?.[0];
            if (!file) return;
            if (!file.type.startsWith('image/')) {
                this.errormsg = 'Please upload a valid image file.';
                return;
            }
            const max = 10 * 1024 * 1024;
            if (file.size > max) { this.errormsg = 'Image too large (max 10MB)'; if (e?.target) e.target.value=''; return; }
            const reader = new FileReader();
            reader.onload = () => {
                this.newPhoto = reader.result;
                this.errormsg = null;
            };
            reader.onerror = () => {
                this.errormsg = 'Error reading file.';
            };
            reader.readAsDataURL(file);
        },
        async updatePhoto() {
            if (!this.newPhoto) return
            try {
                // send base64 payload only (no data URL prefix)
                const b64 = (this.newPhoto.includes(',') ? this.newPhoto.split(',')[1] : this.newPhoto) || '';
                await axios.put('/user/photo', { photo: b64 });
                this.user.photo = this.newPhoto
                localStorage.setItem('userPhoto', this.user.photo)
                this.newPhoto = null
                this.errormsg = null
            } catch (error) {
                console.error('Error updating photo:', error);
                this.errormsg = error.response?.data?.message || 'Failed to update photo.';
            }
        },
    },
    mounted() {
        this.initFromLocal();
    },
}
</script>
<style scoped>
/* profile */
.profile {
  min-height: 100vh;
  background: var(--bg);
  color: var(--text);
  padding: 1rem;
  display: grid;
  gap: 1rem;
}

/* header */
.profile-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: .75rem;
  border-bottom: 1px solid var(--border);
  padding-bottom: .5rem;
}
.title {
  margin: 0;
  font-size: clamp(1.1rem, 1rem + .6vw, 1.5rem);
  font-weight: 700;
}
.hello { color: var(--text); }
.muted { color: var(--text-dim); }
.actions { display: flex; gap: .5rem; }

/* container */
.profile-container {
  background: var(--bg-alt);
  border: 1px solid var(--border);
  border-radius: 14px;
  padding: 1rem;
  box-shadow: 0 10px 30px rgba(0,0,0,.35);
}

.profile-head {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 1rem;
  align-items: center;
}
.photo-wrap { display: grid; place-items: center; }
.avatar {
  width: 100px; height: 100px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid var(--accent-alt);
  box-shadow: 0 0 10px var(--accent-alt);
  background: var(--bg);
}
.no-photo {
  width: 100px; height: 100px;
  display: grid; place-items: center;
  border-radius: 50%;
  border: 1px dashed var(--border);
  color: var(--text-dim);
  background: var(--bg);
}
.identity { display: grid; gap: .6rem; }
.username {
  margin: 0;
  font-size: clamp(1.2rem, 1rem + .8vw, 1.6rem);
  font-weight: 800;
}

/* edit */
.editor { display: grid; gap: .6rem; }
.field-row { display: flex; gap: .5rem; flex-wrap: wrap; }
.input {
  flex: 1 1 240px;
  padding: .6rem .7rem;
  border-radius: var(--radius);
  background: var(--bg-hover);
  border: 1px solid var(--border);
  color: var(--text);
  transition: .15s ease;
}
.input::placeholder { color: var(--text-dim); }
.input:focus {
  outline: none;
  border-color: var(--accent);
  box-shadow: 0 0 8px var(--accent);
  background: #2d2d2d;
}

/* buttons */
.btn {
  appearance: none;
  background: color-mix(in oklab, var(--bg) 85%, var(--accent) 15%);
  border: 1px solid var(--border);
  color: var(--text);
  padding: .55rem .8rem;
  border-radius: var(--radius);
  cursor: pointer;
  transition: background .15s ease, transform .06s ease, box-shadow .15s ease;
}
.btn:hover { background: var(--bg-hover); transform: translateY(-1px); }
.btn:active { transform: translateY(0); }
.btn:disabled { opacity: .6; cursor: not-allowed; }

.btn-primary {
  background: linear-gradient(90deg, var(--accent), var(--accent-alt));
  color: #0b0f17;
  border: 1px solid color-mix(in oklab, var(--accent) 60%, var(--accent-alt) 40%);
  box-shadow: 0 0 10px var(--accent), 0 0 16px var(--accent-alt);
}
.btn-primary:hover { filter: saturate(1.05) brightness(1.03); }

.btn-danger {
  background: color-mix(in oklab, var(--bg) 80%, #ef4444 20%);
  border-color: color-mix(in oklab, var(--border) 70%, #ef4444 30%);
  box-shadow: 0 0 8px rgba(239, 68, 68, .35);
}
.btn-danger:hover { background: color-mix(in oklab, var(--bg) 70%, #ef4444 30%); }

/* responsive */
@media (max-width: 640px) {
  .profile-top { flex-direction: column; align-items: flex-start; gap: .5rem; }
}
</style>
