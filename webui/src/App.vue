<script setup>
import { RouterLink, RouterView } from 'vue-router'
</script>

<template>
  <div class="app">
    <!-- topbar -->
    <header class="app-header">
      <RouterLink to="/" class="logo">WASAText</RouterLink>
    </header>

    <!-- layout -->
    <div class="app-body">
      <!-- sidebar -->
      <aside class="app-sidebar">
        <nav>
          <h6>General</h6>
          <ul>
            <li><RouterLink to="/home">My Conversations</RouterLink></li>
            <li><RouterLink to="/profile">My Profile</RouterLink></li>
            <li><RouterLink to="/search">Search People/Chats</RouterLink></li>
          </ul>

          <h6>Groups</h6>
          <ul>
            <li><RouterLink to="/groups">My Groups</RouterLink></li>
          </ul>
        </nav>
      </aside>

      <!-- main slot -->
      <main class="app-main">
        <RouterView />
      </main>
    </div>
  </div>
</template>

<style>

.app {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}
.app-header {
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--border);
}
.logo {
  font-weight: 700;
  text-decoration: none;
  color: var(--text);
}
.app-body {
  flex: 1;
  display: flex;
}
.app-sidebar {
  width: 220px;
  border-right: 1px solid var(--border);
  padding: 1rem;
}
.app-sidebar h6 {
  margin: 1rem 0 0.5rem;
  font-size: 0.8rem;
  color: var(--text-dim);
  text-transform: uppercase;
}
.app-sidebar ul {
  list-style: none;
  margin: 0;
  padding: 0;
}
.app-sidebar li {
  margin: 0.25rem 0;
}
.app-sidebar a {
  text-decoration: none;
  color: var(--text);
}
.app-main {
  flex: 1;
  padding: 1rem;
}
</style>
