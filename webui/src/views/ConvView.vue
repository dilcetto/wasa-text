<template>
  <section class="conversation">
<header class="conv-header">
      <router-link to="/home" class="back">← Back</router-link>
  <div class="peer">
    <img v-if="conversationPhoto" :src="conversationPhoto" class="avatar" :alt="(conversation.displayName || 'Chat') + ' photo'" />
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
      <div v-if="toast.show" class="toast">{{ toast.msg }}
        <button v-if="toast.targetId" class="link" @click="openToastTarget">Open chat</button>
      </div>

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
              <img :src="'data:image/png;base64,' + m.attachments[0]" alt="Image attachment" />
            </div>

            <div class="meta-line">
              <span class="sender">{{ m.sender?.username || 'Unknown' }}</span>
              <span class="time">{{ formatTime(m.timestamp) }}</span>
              <span class="status" v-if="isOwn(m) && m.message_status" :class="statusClass(m.message_status)" :title="m.message_status">{{ statusIcon(m.message_status) }}</span>
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
              <button type="button" class="link" @click.stop.prevent="openReply(m)">Reply</button>
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
        <div v-if="reply.active" class="reply-banner">
          <span class="label">Replying to {{ reply.username || 'message' }}:</span>
          <span class="snippet">{{ reply.preview }}</span>
          <button class="link" @click="cancelReply">✕</button>
        </div>
        <input
          v-model="newMessage"
          class="input"
          type="text"
          placeholder="Type your message..."
          @keyup.enter="send"
          ref="messageInput"
        />
        <input ref="fileInput" type="file" accept="image/*" @click="resetFileInput" @change="attachPhoto" aria-label="Attach image" />
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
            <button class="btn" :disabled="!forward.target || forwarding" @click="doForward">{{ forwarding ? 'Forwarding…' : 'Forward' }}</button>
            <button class="btn secondary" @click="closeForward">Cancel</button>
          </div>

          <div class="divider">or</div>
          <div class="new-chat">
            <div class="new-chat-input">
              <input v-model="forward.newUsername" class="input" placeholder="Username (exact)" @input="onForwardUserInput" />
              <div v-if="forward.suggestLoading" class="hint muted">Searching…</div>
              <div v-else-if="forward.newUsername && !forward.suggestions.length" class="hint muted">No matches</div>
              <div v-if="forward.suggestions && forward.suggestions.length" class="suggestions">
                <div class="item" v-for="u in forward.suggestions" :key="u.id" @click="selectForwardSuggestion(u)">@{{ u.username }}</div>
              </div>
            </div>
            <button class="btn" :disabled="!canForwardNew || forwarding" @click="doForwardNew">{{ forwarding ? 'Forwarding…' : 'Create & Forward' }}</button>
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
            toast: { show: false, msg: "", targetId: '' },
            reply: { active: false, preview: '', username: '' },
            conversation: {
                id: null,
                displayName: '',
                photo: null,
                membersIds: [],
                messages: []
            },
            allConversations: [],
            forward: { open: false, messageId: null, target: '', newUsername: '', suggestions: [], suggestTimer: null, suggestLoading: false, selectedUserId: '' },
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
            return b64 ? 'data:image/png;base64,' + b64 : 'nopfp.jpg';
        },
        canSend() {
            // Allow sending if we have text or an image attachment
            return (this.newMessage && this.newMessage.trim().length > 0) || !!this.photoAttachB64;
        },
        canForwardNew() {
            const v = (this.forward?.newUsername || '').trim();
            const okLen = v.length >= 3 && v.length <= 16;
            const okChars = /^[A-Za-z0-9_]+$/.test(v);
            return okLen && okChars; // matches Username constraints
        },
    },
methods: {
    onKeyDown(e) {
      if (e?.key === 'Escape' && this.forward.open) this.closeForward();
    },
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
            const token = localStorage.getItem('token');
            const response = await this.$axios.get(`/conversations/${this.conversationId}`, token ? { headers: { Authorization: `Bearer ${token}` } } : {});
            // backend return message.content as base64
            this.conversation = response.data || {};
            this.$nextTick(() => {
              this.scrollToBottom();
              this.markRead();
            });
        } catch (error) {
            console.error('Error loading conversation:', error);
            this.errorMessage = 'Failed to load conversation';
        } finally {
            this.loading = false;
        }
    },
    async loadConversationsList() {
        try {
            const token = localStorage.getItem('token');
            const res = await this.$axios.get('/conversations', token ? { headers: { Authorization: `Bearer ${token}` } } : {});
            this.allConversations = res.data || [];
      } catch (e) {
        console.error('Failed to load conversations list', e);
      }
    },
    showToast(msg) {
      this.toast = { show: true, msg, targetId: this.toast?.targetId || '' };
      setTimeout(() => { this.toast.show = false; }, 2000);
    },
    openToastTarget() {
      const id = this.toast?.targetId;
      if (id) this.$router.push(`/conversations/${id}`);
      this.toast = { show: false, msg: "", targetId: '' };
    },
    async markRead() {
      try {
        const token = localStorage.getItem('token');
        const headers = token ? { Authorization: `Bearer ${token}` } : {};
        const msgs = (this.conversation?.messages || []).filter(m => m && m.id && m.senderId !== this.userId);
        for (const m of msgs) {
          try {
            await this.$axios.post(`/conversations/${this.conversationId}/messages/${m.id}/status`, { status: 'read' }, { headers });
          } catch (e) {
            // non-blocking; continue
          }
        }
      } catch {}
    },
    async send() {
        if (!this.canSend || this.sending) return;
        this.sending = true;
        this.errorMessage = null;
        try {
            let messagePayload = {};
            let textToSend = this.newMessage || '';
            if (this.reply.active) {
              const quoted = `> ${this.reply.username ? '@' + this.reply.username + ': ' : ''}${this.reply.preview}`;
              textToSend = quoted + (textToSend ? `\n\n${textToSend}` : '');
            }
            messagePayload = {
              content: {
                type: 'text',
                value: textToSend ? this.toBase64(textToSend) : ''
              },
              ...(this.photoAttachB64 ? { attachments: [ this.photoAttachB64 ] } : {})
            };
            const token = localStorage.getItem('token');
            await this.$axios.post(`/conversations/${this.conversationId}/messages`, messagePayload, token ? { headers: { Authorization: `Bearer ${token}` } } : {});
            this.newMessage = '';
            this.photoAttachB64 = '';
            if (this.$refs.fileInput) this.$refs.fileInput.value = '';
            this.reply = { active: false, preview: '', username: '' };
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
      // 10MB limit to match backend
      const max = 10 * 1024 * 1024;
      if (file.size > max) { this.errorMessage = 'Image too large (max 10MB)'; if (e?.target) e.target.value=''; return; }
      const reader = new FileReader();
      const inputEl = e?.target || null;
      reader.onload = () => {
        const result = reader.result || '';
        this.photoAttachB64 = typeof result === 'string' && result.includes(',') ? result.split(',')[1] : result;
      };
      reader.onerror = () => { this.errorMessage = 'Failed to read file'; };
      reader.readAsDataURL(file);
    },
    statusIcon(status) {
      const s = String(status || '').toLowerCase();
      if (s === 'read') return '✓✓';
      if (s === 'delivered') return '✓✓';
      return '✓'; // sent
    },
    statusClass(status) {
      const s = String(status || '').toLowerCase();
      return s === 'read' ? 'read' : (s === 'delivered' ? 'delivered' : 'sent');
    },
    resetFileInput(e) {
      if (e?.target) e.target.value = '';
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
    openReply(message) {
        const text = (this.decodeText(message?.content?.value) || '').trim();
        const hasImage = Array.isArray(message?.attachments) && message.attachments.length > 0;

        // reply preview that always shows both types when present.
        // ensure 'Photo' label is preserved even when truncating long text.
        let preview = '';
        if (hasImage && text) {
          const suffix = ' • Photo';
          const max = 80 - suffix.length; // leave room for the suffix
          const truncated = text.length > max ? (text.slice(0, Math.max(0, max - 3)) + '...') : text;
          preview = truncated + suffix;
        } else if (hasImage) {
          preview = 'Photo';
        } else {
          preview = text;
        }

        const username = message?.sender?.username || '';
        this.reply = { active: true, preview, username };
        this.$nextTick(() => this.$refs.messageInput?.focus());
    },
    cancelReply() {
        this.reply = { active: false, preview: '', username: '' };
    },
    closeForward() {
        this.forward = { open: false, messageId: null, target: '', newUsername: '', suggestions: [], suggestTimer: null, suggestLoading: false, selectedUserId: '' };
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
            this.toast = { show: true, msg: "Message forwarded.", targetId: this.forward.target };
            setTimeout(() => { this.toast.show = false; }, 2000);
            this.closeForward();
            // If forwarded to the current conversation, refresh messages
            if (this.forward.target === this.conversationId) {
              await this.load();
            }
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
    async doForwardNew() {
        if (!this.forward.messageId || !this.canForwardNew) return;
        this.forwarding = true;
        try {
            const username = (this.forward.newUsername || '').trim();
            const token = localStorage.getItem('token');
            const headers = token ? { Authorization: `Bearer ${token}` } : {};

            // Prefer selected suggestion; otherwise search for the user
            let userId = this.forward.selectedUserId || '';
            if (!userId) {
              const sr = await this.$axios.get(`/searchby?user=${encodeURIComponent(username)}`, { headers });
              const users = (sr?.data?.users || []).filter(u => u && u.id && (u.username || '').toLowerCase() === username.toLowerCase());
              userId = users?.[0]?.id || '';
            }
            if (!userId) {
                this.errorMessage = 'User not found';
                return;
            }

            // ensure or create the direct conversation
            const cd = await this.$axios.post('/direct-conversations', { peerUserId: userId }, { headers });
            const targetId = cd?.data?.conversationId;
            if (!targetId) {
                this.errorMessage = 'Failed to create conversation';
                return;
            }

            // forward the message to the new conversation
            await this.$axios.post(
              `/conversations/${this.conversationId}/messages/${this.forward.messageId}/forward`,
              { targetConversationId: targetId },
              { headers }
            );
            this.toast = { show: true, msg: 'Message forwarded.', targetId };
            setTimeout(() => { this.toast.show = false; }, 2000);
            this.closeForward();
        } catch (e) {
            console.error('Failed to forward to new user', e);
            this.errorMessage = 'Failed to forward';
        } finally {
            this.forwarding = false;
            this.$nextTick(() => { this.$refs.messageInput?.focus(); });
        }
    },
    onForwardUserInput() {
      // reset selected user and debounce search
      this.forward.selectedUserId = '';
      const q = (this.forward.newUsername || '').trim();
      if (this.forward.suggestTimer) clearTimeout(this.forward.suggestTimer);
      if (q.length < 2) { this.forward.suggestions = []; this.forward.suggestLoading = false; return; }
      this.forward.suggestLoading = true;
      this.forward.suggestTimer = setTimeout(async () => {
        try {
          const token = localStorage.getItem('token');
          const headers = token ? { Authorization: `Bearer ${token}` } : {};
          const res = await this.$axios.get(`/searchby?user=${encodeURIComponent(q)}`, { headers });
          const all = res?.data?.users || [];
          // Exclude self and limit results
          this.forward.suggestions = all.filter(u => u && u.id && u.id !== this.userId).slice(0, 6);
        } catch { this.forward.suggestions = []; }
        finally { this.forward.suggestLoading = false; }
      }, 250);
    },
    selectForwardSuggestion(u) {
      if (!u) return;
      this.forward.newUsername = u.username || '';
      this.forward.selectedUserId = u.id || '';
      this.forward.suggestions = [];
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
    try { window.addEventListener('keydown', this.onKeyDown); } catch {}
    this.pollId = setInterval(() => this.load(), 10000);
    this.$nextTick(() => {
      this.$refs.messageInput?.focus();
    });
  },
  unmounted() {
    if (this.pollId) clearInterval(this.pollId);
    try { window.removeEventListener('keydown', this.onKeyDown); } catch {}
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
.status { margin-left: .25rem; }
.status.sent { color: var(--text-dim); }
.status.delivered { color: var(--text-dim); }
.status.read { color: var(--accent-alt); text-shadow: 0 0 6px var(--accent-alt); }
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
.reply-banner { display: flex; align-items: center; gap: .5rem; margin-right: auto; color: var(--text-dim); }
.reply-banner .label { font-weight: 600; }
.reply-banner .snippet { max-width: 32ch; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
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
.forward-overlay { position: fixed; inset: 0; background: rgba(0,0,0,.45); display: grid; place-items: center; z-index: 2000; }
.forward-card { background: var(--bg); border: 1px solid var(--border); border-radius: 12px; padding: 1rem; width: min(92vw, 420px); color: var(--text); box-shadow: var(--shadow); display: grid; gap: .6rem; z-index: 2001; }
.select { width: 100%; padding: .5rem; border: 1px solid var(--border); border-radius: var(--radius); background: var(--bg-alt); color: var(--text); }
.forward-actions { display: flex; gap: .5rem; justify-content: flex-end; }
.btn.secondary { background: var(--bg-alt); color: var(--text); border: 1px solid var(--border); }
.divider { text-align: center; color: var(--text-dim); margin-top: .25rem; }
.new-chat { display: flex; gap: .5rem; align-items: flex-start; }
.new-chat-input { position: relative; flex: 1; }
.suggestions { position: absolute; top: 100%; left: 0; right: 0; background: var(--bg); border: 1px solid var(--border); border-radius: 8px; margin-top: .25rem; max-height: 200px; overflow: auto; z-index: 10; }
.suggestions .item { padding: .4rem .6rem; cursor: pointer; }
.suggestions .item:hover { background: var(--bg-hover); }
.hint { font-size: .8rem; margin-top: .25rem; }

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
