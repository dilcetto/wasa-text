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

            <div v-if="m.content?.type === 'text'" class="text" v-text="decodeText(m.content?.value)"></div>
            <div v-else-if="m.content?.type === 'image'" class="image">
              <img :src="'data:image/png;base64,' + (m.content?.value || '')" alt="image" />
            </div>
            <div v-else-if="m.attachments?.length" class="attachment">
              <img :src="'data:image/png;base64,' + m.attachments[0]" alt="attachment" />
            </div>

            <div class="meta-line">
              <span class="sender">{{ m.sender?.username || 'Unknown' }}</span>
              <span class="time">{{ formatTime(m.timestamp) }}</span>
              <span class="status" v-if="m.message_status">• {{ m.message_status }}</span>
            </div>

            <div class="actions">
              <button class="link" @click="openForward(m.id)">Forward</button>
              <button v-if="isOwn(m)" class="link danger" @click="del(m.id)">Delete</button>
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
            userId: localStorage.getItem('userId') || null
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
            return this.newMessage.trim().length > 0;
        },
    },
methods: {
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
            const messagePayload = {
                content: {
                    type: 'text',
                    value: this.toBase64(this.newMessage)
                },
            };
            await this.$axios.post(`/conversations/${this.conversationId}/messages`, messagePayload);
            this.newMessage = '';
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
   async del(messageId) {
    if (!messageId) return;
    if (!confirm("Are you sure you want to delete this message?")) return;
    try {
        await this.$axios.delete(`/conversations/${this.conversationId}/messages/${messageId}`);
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
            await this.$axios.post(`/conversations/${this.conversationId}/messages/${this.forward.messageId}/forward`, { targetConversationId: this.forward.target });
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
  border: 1px solid var(--border);
  border-radius: 12px;
}
.conv-header {
  display: flex;
  align-items: center;
  gap: .75rem;
  padding: .75rem 1rem;
  border-bottom: 1px solid var(--border);
}
.back { text-decoration: none; color: var(--text-dim); }
.peer { display: flex; align-items: center; gap: .75rem; }
.avatar { width: 36px; height: 36px; border-radius: 50%; object-fit: cover; }
.meta { display: grid; }
.name { margin: 0; font-size: 1rem; }
.muted { color: var(--text-dim); font-size: .85rem; }
.messages {
  overflow-y: auto;
  padding: 1rem;
  display: flex;
  flex-direction: column;
  gap: .5rem;
  background: var(--bg);
}
.msg-row { display: flex; }
.msg-row.own { justify-content: flex-end; }
.bubble {
  max-width: 70ch;
  padding: .6rem .7rem;
  border-radius: 12px;
  background: var(--bg-alt);
  border: 1px solid var(--border);
  box-shadow: var(--shadow);
}
.msg-row.own .bubble {
  background: color-mix(in oklab, var(--bg-alt) 70%, var(--accent) 30%);
  border-color: color-mix(in oklab, var(--border) 40%, var(--accent) 60%);
}
.fwd { font-size: .75rem; color: var(--text-dim); margin-bottom: .25rem; }
.text { white-space: pre-wrap; word-break: break-word; }
.image img, .attachment img { max-width: 420px; border-radius: 8px; display: block; }
.meta-line { margin-top: .35rem; font-size: .75rem; color: var(--text-dim); display: flex; gap: .35rem; }
.actions { margin-top: .35rem; display: flex; gap: .5rem; }
.link { background: transparent; border: none; color: var(--accent); cursor: pointer; padding: 0; }
.link.danger { color: #ef4444; }
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
