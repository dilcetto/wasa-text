<template>
  <section class="conversation">
<header class="conv-header">
      <router-link to="/home" class="back">← Back</router-link>
  <div class="peer">
    <img v-if="conversationPhoto" :src="conversationPhoto" class="avatar" alt="Chat" />
    <div class="meta">
      <h2 class="name">{{ conversation.displayName || 'Conversation' }}</h2>
      <div class="muted" v-if="conversation.membersIds?.length">
        {{ conversation.membersIds.length }} member(s)
      </div>
      <div v-if="conversation.type === 'group'" class="actions">
        <router-link :to="`/groups/${conversationId}/edit`" class="link">Edit Group</router-link>
      </div>
    </div>
  </div>
    </header>

    <LoadingSpinner :loading="loading">
      <ErrorMsg v-if="errorMessage" :msg="errorMessage" />
      <div v-if="toast.show" class="toast">{{ toast.msg }}</div>

      <div ref="scrollArea" class="messages">
        <div
          v-for="m in conversation.messages || []"
          :key="m.id"
          class="msg-row"
          :class="{ own: isOwn(m) }"
        >
          <div class="bubble">
            <div v-if="m.forwarded_from" class="fwd">Forwarded</div>

            <div v-if="decodeText(m.content?.value)" class="text" v-text="decodeText(m.content?.value)"></div>
            <div v-if="m.attachments?.length" class="attachment">
              <img :src="'data:image/png;base64,' + m.attachments[0]" alt="attachment" />
            </div>

            <div class="meta-line">
              <span class="sender">{{ m.sender?.username || 'Unknown' }}</span>
              <span class="time">{{ formatTime(m.timestamp) }}</span>
              <span class="status" v-if="m.message_status">• {{ m.message_status }}</span>
            </div>

            <div class="reactions" v-if="m.reaction && m.reaction.length">
              <span
                v-for="g in groupReactions(m)"
                :key="g.emoji"
                class="rx-chip"
                :class="{ mine: g.mine }"
                :title="g.users.join(', ')"
              >
                {{ g.emoji }} {{ g.count }}
              </span>
            </div>

            <div class="actions">
              <button type="button" class="link" @click.stop.prevent="openForward(m.id)">Forward</button>
              <button v-if="isOwn(m)" type="button" class="link danger" @click.stop.prevent="del(m.id)">Delete</button>
              <span class="sep">|</span>
              <span class="muted">React:</span>
              <button
                type="button"
                class="link emoji"
                :class="{ active: myReaction(m) === '👍', busy: !!reactBusy[m.id] }"
                :disabled="!!reactBusy[m.id]"
                @click.stop.prevent="react(m.id, '👍')"
              >👍</button>
              <button
                type="button"
                class="link emoji"
                :class="{ active: myReaction(m) === '❤️', busy: !!reactBusy[m.id] }"
                :disabled="!!reactBusy[m.id]"
                @click.stop.prevent="react(m.id, '❤️')"
              >❤️</button>
              <button
                type="button"
                class="link emoji"
                :class="{ active: myReaction(m) === '😂', busy: !!reactBusy[m.id] }"
                :disabled="!!reactBusy[m.id]"
                @click.stop.prevent="react(m.id, '😂')"
              >😂</button>
              <button
                type="button"
                class="link"
                :disabled="!!reactBusy[m.id] || !myReaction(m)"
                @click.stop.prevent="unreact(m.id)"
              >Remove</button>
            </div>
          </div>
        </div>
      </div>

      <footer class="composer">
        <input
          v-model="newMessage"
          class="input"
          type="text"
          placeholder="Type your message..."
          @keyup.enter="send"
          ref="messageInput"
        />
        <input type="file" accept="image/*" @change="attachPhoto" />
        <button class="btn" :disabled="!canSend || sending" @click="send">{{ sending ? 'Sending…' : 'Send' }}</button>
      </footer>

      <div v-if="forward.open" class="forward-overlay" @click.self="closeForward">
        <div class="forward-card">
          <h3>Forward message</h3>
          <select v-model="forward.target" class="select">
            <option disabled value="">Select a conversation</option>
            <option v-for="c in allConversations" :key="c.conversationId" :value="c.conversationId">
              {{ c.displayName }}
            </option>
          </select>
          <div class="forward-actions">
            <button class="btn" :disabled="!forward.target || forward.target === conversationId || forwarding" @click="doForward">{{ forwarding ? 'Forwarding…' : 'Forward' }}</button>
            <button class="btn secondary" @click="closeForward">Cancel</button>
          </div>
        </div>
      </div>
    </LoadingSpinner>
  </section>
</template>

<script>
export default {
    name: "ConvView",
    data() {
        return {
            loading: true,
            sending: false,
            forwarding: false,
            errorMessage: null,
            newMessage: '',
            photoAttachB64: '',
            toast: { show: false, msg: "" },
            conversation: {
                id: null,
                displayName: '',
                photo: null,
                membersIds: [],
                messages: []
            },
            allConversations: [],
            forward: { open: false, messageId: null, target: '' },
            pollId: null,   
            userId: localStorage.getItem('userId') || null,
            reactBusy: {},
        };
    },
    computed: {
        conversationId() {
            return this.$route.params.conversationId;
        },
        conversationPhoto() {
            const b64 = this.conversation?.profilePhoto || this.conversation?.photo;
            return b64 ? 'data:image/png;base64,' + b64 : '';
        },
        canSend() {
            // Allow sending if we have text or an image attachment
            return (this.newMessage && this.newMessage.trim().length > 0) || !!this.photoAttachB64;
        },
    },
methods: {
    myReaction(message) {
      try {
        const uid = this.userId;
        const arr = message?.reaction || [];
        const mine = arr.find(r => r.user_id === uid || r.userId === uid);
        return mine?.emoji || '';
      } catch { return ''; }
    },
    groupReactions(message) {
      try {
        const arr = message?.reaction || [];
        const map = new Map();
        for (const r of arr) {
          const emoji = r?.emoji || '';
          const uid = r?.user_id || r?.userId || '';
          const uname = r?.username || '';
          if (!emoji) continue;
          if (!map.has(emoji)) map.set(emoji, { emoji, users: [], count: 0, mine: false });
          const g = map.get(emoji);
          g.users.push(uname || uid);
          g.count += 1;
          if (uid === this.userId) g.mine = true;
        }
        return Array.from(map.values());
      } catch {
        return [];
      }
    },
    async load() {
        this.errorMessage = null;
        try {
            const response = await this.$axios.get(`/conversations/${this.conversationId}`);
            // backend return message.content as base64
            this.conversation = response.data || {};
            this.$nextTick(this.scrollToBottom);
        } catch (error) {
            console.error('Error loading conversation:', error);
            this.errorMessage = 'Failed to load conversation';
        } finally {
            this.loading = false;
        }
    },
    async loadConversationsList() {
        try {
            const res = await this.$axios.get('/conversations');
            this.allConversations = res.data || [];
      } catch (e) {
        console.error('Failed to load conversations list', e);
      }
    },
    showToast(msg) {
      this.toast = { show: true, msg };
      setTimeout(() => { this.toast.show = false; }, 2000);
    },
    async send() {
        if (!this.canSend || this.sending) return;
        this.sending = true;
        this.errorMessage = null;
        try {
            let messagePayload = {};
            messagePayload = {
              content: {
                type: 'text',
                value: this.newMessage ? this.toBase64(this.newMessage) : ''
              },
              ...(this.photoAttachB64 ? { attachments: [ this.photoAttachB64 ] } : {})
            };
            const token = localStorage.getItem('token');
            await this.$axios.post(`/conversations/${this.conversationId}/messages`, messagePayload, token ? { headers: { Authorization: `Bearer ${token}` } } : {});
            this.newMessage = '';
            this.photoAttachB64 = '';
            this.showToast("Message sent.");
            await this.load();
        } catch (error) {
            console.error('Error sending message:', error);
            this.errorMessage = 'Failed to send message';
        } finally {
            this.sending = false;
        }
        this.$nextTick(() => {
          this.$refs.messageInput?.focus();
        });
    },
    attachPhoto(e) {
      const file = e?.target?.files?.[0];
      if (!file) { this.photoAttachB64 = ''; return; }
      if (!file.type.startsWith('image/')) { this.errorMessage = 'Please select an image'; return; }
      const reader = new FileReader();
      reader.onload = () => {
        const result = reader.result || '';
        this.photoAttachB64 = typeof result === 'string' && result.includes(',') ? result.split(',')[1] : result;
      };
      reader.onerror = () => { this.errorMessage = 'Failed to read file'; };
      reader.readAsDataURL(file);
    },
    async react(messageId, emoji) {
      if (this.reactBusy[messageId]) return;
      this.$set ? this.$set(this.reactBusy, messageId, true) : (this.reactBusy[messageId] = true);
      try {
        const token = localStorage.getItem('token');
        const url = `/conversations/${this.conversationId}/messages/${messageId}/comment`;
        if (this.$axios && this.$axios.post) {
          await this.$axios.post(
            url,
            { emoji },
            token ? { headers: { Authorization: `Bearer ${token}` } } : {}
          );
        } else {
          // Fallback to fetch if axios instance is not available in this context
          const res = await fetch((typeof __API_URL__ !== 'undefined' ? __API_URL__ : '') + url, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              ...(token ? { Authorization: `Bearer ${token}` } : {}),
            },
            body: JSON.stringify({ emoji }),
          });
          if (!res.ok) throw new Error(`HTTP ${res.status}`);
        }
        // Optimistically update UI
        const uname = localStorage.getItem('username') || '';
        const msg = (this.conversation?.messages || []).find(m => m.id === messageId);
        if (msg) {
          msg.reaction = (msg.reaction || []).filter(r => (r.user_id || r.userId) !== this.userId);
          msg.reaction.push({ message_id: messageId, user_id: this.userId, emoji, username: uname });
        }
        this.showToast('Reacted');
        // background refresh
        this.load();
      } catch (e) {
        console.error('Failed reaction', e);
        this.errorMessage = 'Failed to react';
      } finally {
        this.reactBusy[messageId] = false;
      }
    },
    async unreact(messageId) {
      if (this.reactBusy[messageId]) return;
      this.$set ? this.$set(this.reactBusy, messageId, true) : (this.reactBusy[messageId] = true);
      try {
        const token = localStorage.getItem('token');
        const url = `/conversations/${this.conversationId}/messages/${messageId}/comment`;
        if (this.$axios && this.$axios.delete) {
          await this.$axios.delete(
            url,
            { data: {}, ...(token ? { headers: { Authorization: `Bearer ${token}` } } : {}) }
          );
        } else {
          const res = await fetch((typeof __API_URL__ !== 'undefined' ? __API_URL__ : '') + url, {
            method: 'DELETE',
            headers: {
              'Content-Type': 'application/json',
              ...(token ? { Authorization: `Bearer ${token}` } : {}),
            },
            body: JSON.stringify({}),
          });
          if (!res.ok) throw new Error(`HTTP ${res.status}`);
        }
        // Optimistically update UI
        const msg = (this.conversation?.messages || []).find(m => m.id === messageId);
        if (msg && Array.isArray(msg.reaction)) {
          msg.reaction = msg.reaction.filter(r => (r.user_id || r.userId) !== this.userId);
        }
        this.showToast('Removed reaction');
        this.load();
      } catch (e) {
        console.error('Failed to remove reaction', e);
        this.errorMessage = 'Failed to remove reaction';
      } finally {
        this.reactBusy[messageId] = false;
      }
    },
   async del(messageId) {
    if (!messageId) return;
    if (!confirm("Are you sure you want to delete this message?")) return;
    try {
        const token = localStorage.getItem('token');
        await this.$axios.delete(`/conversations/${this.conversationId}/messages/${messageId}`,
          token ? { headers: { Authorization: `Bearer ${token}` } } : {}
        );
        this.showToast("Message deleted.");
        await this.load();
      } catch (e) {
        console.error('Failed to delete message', e);
        this.errorMessage = 'Failed to delete message';
      }
    },

    openForward(messageId) {
        this.forward = { open: true, messageId, target: '' };
        if (!this.allConversations?.length) this.loadConversationsList();
    },
    closeForward() {
        this.forward = { open: false, messageId: null, target: '' };
    },
    async doForward() {
        if (!this.forward.messageId || !this.forward.target) return;
        this.forwarding = true;
        try {
            const token = localStorage.getItem('token');
            await this.$axios.post(
              `/conversations/${this.conversationId}/messages/${this.forward.messageId}/forward`,
              { targetConversationId: this.forward.target },
              token ? { headers: { Authorization: `Bearer ${token}` } } : {}
            );
            this.showToast("Message forwarded.");
            this.closeForward();
        } catch (e) {
            console.error('Failed to forward message', e);
            this.errorMessage = 'Failed to forward message';
        } finally {
            this.forwarding = false;
        this.$nextTick(() => {
          this.$refs.messageInput?.focus();
        });
        }
    },
    isOwn(message) {
        return message?.senderId === this.userId;
    },
    formatTime(timestamp) {
        if (!timestamp) return '';
      try {
        return new Date(timestamp).toLocaleString();
      } catch { return timestamp; }
    },
    toBase64(str) {
        const bytes = new TextEncoder().encode(str);
      let binary = '';
      bytes.forEach(b => binary += String.fromCharCode(b));
      return btoa(binary);
    },
    decodeText(b64) {
      if (!b64) return '';
      try {
        const bin = atob(b64);
        const bytes = Uint8Array.from(bin, c => c.charCodeAt(0));
        return new TextDecoder().decode(bytes);
      } catch { return ''; }
    },
    scrollToBottom() {
      const el = this.$refs.scrollArea;
      if (el) el.scrollTop = el.scrollHeight;
    },
  },
  mounted() {
    this.load();
    this.pollId = setInterval(() => this.load(), 10000);
    this.$nextTick(() => {
      this.$refs.messageInput?.focus();
    });
  },
  unmounted() {
    if (this.pollId) clearInterval(this.pollId);
  },
};
</script>

<style scoped>
.conversation {
  display: grid;
  grid-template-rows: auto 1fr auto;
  height: calc(100vh - 100px);
  background: var(--bg);
  color: var(--text);
}
.conv-header {
  display: flex;
  align-items: center;
  gap: .75rem;
  padding: .75rem 1rem;
  border-bottom: 1px solid var(--border);
}
.back { text-decoration: none; color: var(--text-dim); }
.peer { display: flex; flex-direction: column; align-items: center; gap: .4rem; padding-left: .5rem; padding-right: .5rem; }
.avatar { width: 48px; height: 48px; border-radius: 50%; object-fit: cover; border: 2px solid var(--accent); box-shadow: 0 0 6px var(--accent); }
.meta { display: grid; text-align: center; }
.name { margin: 0; font-size: 1rem; }
.muted { color: var(--text-dim); font-size: .85rem; }
.messages {
  overflow-y: auto;
  padding: 1rem;
  display: flex;
  flex-direction: column;
  gap: .5rem;
  background: var(--bg);
  position: relative;
  z-index: 1;
}
.msg-row { display: flex; }
.msg-row.own { justify-content: flex-end; }
.bubble {
  max-width: 70ch;
  padding: .6rem .7rem;
  border-radius: 12px;
  background: var(--bg-alt);
  border: 1px solid var(--border);
  box-shadow: 0 0 10px rgba(0,0,0,.25);
  position: relative;
  z-index: 2;
}
.msg-row.own .bubble {
  background: color-mix(in oklab, var(--bg-alt) 65%, var(--accent) 35%);
  border-color: color-mix(in oklab, var(--border) 30%, var(--accent) 70%);
}
.fwd { font-size: .75rem; color: var(--text-dim); margin-bottom: .25rem; }
.text { white-space: pre-wrap; word-break: break-word; }
.image img, .attachment img { max-width: 420px; border-radius: 8px; display: block; }
.meta-line { margin-top: .35rem; font-size: .75rem; color: var(--text-dim); display: flex; gap: .35rem; }
.actions { margin-top: .35rem; display: flex; gap: .5rem; }
.link { background: transparent; border: none; color: var(--accent); cursor: pointer; padding: 0; }
.link.danger { color: #ef4444; }
/* ensure buttons are clickable above any overlay */
.actions, .actions * { pointer-events: auto; }

/* emoji visual feedback */
.emoji { transition: transform .12s ease, opacity .12s ease, color .12s ease; }
.emoji.active { transform: scale(1.1); color: var(--accent-alt); text-shadow: 0 0 8px var(--accent-alt); }
.emoji.busy { opacity: .6; pointer-events: none; }
.reactions { margin-top: .25rem; display: flex; gap: .35rem; flex-wrap: wrap; }
.rx-chip { font-size: .85rem; background: var(--bg); border: 1px solid var(--border); border-radius: 12px; padding: .1rem .4rem; color: var(--text-dim); }
.rx-chip.mine { border-color: var(--accent-alt); color: var(--accent-alt); box-shadow: 0 0 6px var(--accent-alt); }
.composer {
  display: flex;
  gap: .5rem;
  padding: .75rem;
  border-top: 1px solid var(--border);
  background: var(--bg-alt);
}
.input {
  flex: 1;
  padding: .6rem .7rem;
  border-radius: var(--radius);
  border: 1px solid var(--border);
  background: var(--bg);
  color: var(--text);
}
.btn {
  padding: .55rem .9rem;
  border-radius: var(--radius);
  border: none;
  background: var(--accent);
  color: #000;
  font-weight: 700;
}
.btn:disabled { opacity: .6; cursor: not-allowed; }

/* forward dialog */
.forward-overlay { position: fixed; inset: 0; background: rgba(0,0,0,.45); display: grid; place-items: center; }
.forward-card { background: var(--bg); border: 1px solid var(--border); border-radius: 12px; padding: 1rem; width: min(92vw, 420px); color: var(--text); box-shadow: var(--shadow); display: grid; gap: .6rem; }
.select { width: 100%; padding: .5rem; border: 1px solid var(--border); border-radius: var(--radius); background: var(--bg-alt); color: var(--text); }
.forward-actions { display: flex; gap: .5rem; justify-content: flex-end; }
.btn.secondary { background: var(--bg-alt); color: var(--text); border: 1px solid var(--border); }

.toast {
  position: fixed;
  bottom: 2rem;
  left: 50%;
  transform: translateX(-50%);
  background: var(--bg-alt);
  color: var(--accent);
  padding: 0.75rem 1.5rem;
  border-radius: var(--radius);
  box-shadow: var(--shadow);
  font-weight: bold;
  z-index: 1000;
  opacity: 0.95;
  pointer-events: none;
  transition: opacity 0.3s;
}
</style>
