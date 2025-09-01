import {createRouter, createWebHashHistory} from 'vue-router'
import HomeView from '../views/HomeView.vue'
import LoginView from '../views/LoginView.vue'
import ProfileView from '../views/ProfileView.vue'
import ConvView from '../views/ConvView.vue'
import SearchView from '../views/SearchView.vue'
import GroupEditView from '../views/GroupEditView.vue'
import GroupCreateView from '../views/GroupCreateView.vue'

const router = createRouter({
	history: createWebHashHistory(import.meta.env.BASE_URL),
	routes: [
		{path:'/', redirect: '/login'},
		{path: '/login', component: LoginView},
		{path: '/home', component: HomeView},
		{path: '/conversations/:conversationId', component: ConvView},
    {path: '/profile', component: ProfileView},
    {path: '/search', component: SearchView},
    {path: '/groups', component: HomeView},
    {path: '/new-group', component: GroupCreateView},
    {path: '/groups/:groupId/edit', component: GroupEditView},
	]
})

// Minimal global auth guard: protect all routes except /login
router.beforeEach((to, _from, next) => {
  const token = localStorage.getItem('token');
  if (!token && to.path !== '/login') return next('/login');
  next();
});

export default router
