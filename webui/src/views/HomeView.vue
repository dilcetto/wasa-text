<template>
	
  <div class="home-container">
    <div class="sidebar">
      <h2 class="logo">WASAText</h2>
      <input
        v-model="searchQuery"
        class="search-bar"
        type="text"
        placeholder="Search for a user or chat"
      />
      <div class="user-info">
        <img :src="user.photo" alt="User Photo" class="user-photo" />
        <p>{{ user.name }}</p>
      </div>
    </div>
    <div class="chat-list">
      <div
        v-for="chat in filteredChats"
        :key="chat.id"
        class="chat-preview"
        @click="viewConversation(chat.id)"
      >
        <img :src="'data:image/png;base64,' + (chat.photo || '')" alt="Chat Photo" class="chat-photo" />
        <div class="chat-details">
          <h3>{{ chat.name }}</h3>
          <p v-if="chat.lastMessage" class="last-message">
            <span v-if="isForwarded(chat.lastMessage)" v-html="getFormattedMessage(chat.lastMessage)"></span>
            <span v-else>{{ getFormattedMessage(chat.lastMessage) }}</span>
            <span> • {{ new Date(chat.lastMessage.timestamp).toLocaleString() }}</span>
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios';
export default {
  data() {
    return {
      searchQuery: '',
      user: {
        name: localStorage.getItem('username') || 'Unknown',
        photo: '/nopfp.jpg',
      },
      conversations: [],
      errormsg: null,
    };
  },
  methods: {
    async loadConversations() {
      this.errormsg = null;
      try {
        const token = localStorage.getItem("token");
        if (!token) {
          this.$router.push({ path: "/" });
          return;
        }
        const response = await this.$axios.get("/conversations", {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });
        this.conversations = response.data || [];
      } catch (error) {
        console.error("Error loading conversations:", error);
        this.errormsg = "Failed to load conversations.";
      }
    },
    viewConversation(conversationId) {
      this.$router.push(`/conversations/${conversationId}`);
    },
    truncateText(text, length = 50, clamp = '...') {
      if (!text || text.length <= length) return text;
      const lastSpaceIndex = text.substring(0, length).lastIndexOf(' ');
      return lastSpaceIndex === -1 ? text.substring(0, length) + clamp : text.substring(0, lastSpaceIndex) + clamp;
    },
    isForwarded(message) {
      return message.content.includes("<strong>Forwarded from");
    },
    getFormattedMessage(message) {
      if (this.isForwarded(message)) {
        return message.content;
      }
      return this.truncateText(message.content);
    },
  },
  mounted() {
    this.loadConversations();
  },
  computed: {
    filteredChats() {
      const query = this.searchQuery.toLowerCase();
      return this.conversations.filter(
        (chat) =>
          chat.name.toLowerCase().includes(query) ||
          (chat.lastMessage && chat.lastMessage.content.toLowerCase().includes(query))
      );
    },
  },
  created() {
    this.$axios = axios;
  },
};
</script>

<style scoped>
.home-container {
  display: flex;
  height: 100vh;
  background: var(--bg);
  color: var(--text);
}
/*sidebar*/
.sidebar {
  width: 250px;
  padding: 1rem;
  background-color: var(--bg-alt);
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
}


.logo {
  font-size: 1.3rem;
  font-weight: bold;
  margin-bottom: 1rem;
  color: var(--accent);
  text-shadow: 0 0 6px var(--accent);
}

.search-bar {
  padding: 0.5rem;
  border-radius: var(--radius);
  border: none;
  outline: none;
  background: var(--bg-hover);
  color: var(--text);
  margin-bottom: 1rem;
  transition: 0.2s;
}
.search-bar:focus {
  box-shadow: var(--shadow);
  background: #2d2d2d;
}

.user-info {
  margin-top: auto;
  text-align: center;
  padding-top: 1rem;
  border-top: 1px solid var(--border);
  color: var(--text-dim);
}

.user-photo {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  margin-bottom: 0.5rem;
  border: 2px solid var(--accent-alt);
  box-shadow: 0 0 6px var(--accent-alt);
}
/*chat list*/
.chat-list {
  flex: 1;
  overflow-y: auto;
  padding: 1rem;
  background: var(--bg);
}

.chat-preview {
  display: flex;
  align-items: center;
  padding: 0.6rem;
  margin-bottom: 0.6rem;
  border-radius: var(--radius);
  cursor: pointer;
  transition: background-color 0.2s, transform 0.1s;
}

.chat-preview:hover {
  background-color: var(--bg-hover);
  transform: translateX(2px);
}

.chat-photo {
  width: 45px;
  height: 45px;
  border-radius: 50%;
  margin-right: 0.8rem;
  border: 2px solid var(--accent);
  box-shadow: 0 0 6px var(--accent);
}

.chat-details h3 {
  margin: 0;
  font-size: 1rem;
  color: var(--text);
}

.chat-details .last-message {
  margin: 0;
  font-size: 0.85rem;
  color: var(--text-dim);
}

/*scrollbar styling*/
.chat-list::-webkit-scrollbar {
  width: 8px;
}
.chat-list::-webkit-scrollbar-thumb {
  background: var(--accent);
  border-radius: 4px;
}
.chat-list::-webkit-scrollbar-track {
  background: var(--bg-alt);
}
</style>





