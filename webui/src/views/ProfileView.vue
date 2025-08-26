<template>
    <div>
        <div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
            <h1 class="h2">{{ username }}, <span class="text-muted">Welcome to your profile</span></h1>
            <div class="btn-toolbar mb-2 mb-md-0">
                <div class="btn-group me-2">
                    <button type="button" class="btn btn-sm btn-outline-secondary" @click="editProfile = !editProfile">Edit Profile</button>
                    <button type="button" class="btn btn-sm btn-outline-secondary" @click="logOut">Log Out</button>
                </div>
                <div class="btn-group me-2">
                    <button type="button" class="btn btn-sm btn-outline-secondary" @click="startChat">Start Chat</button>
            </div>
        </div>
    </div>

    <div class="profile-container">
        <div class="profile-header">
            <div class="profile-photo">
                <img v-if="user.photo" :src="user.photo" alt="Profile Photo" class="rounded-circle" width="100" height="100"/>
                <p v-else>No Photo Available</p>
            </div>
            <div class="username-container">
                <h1 class="username">{{ username }}</h1>
                <div v-if="editProfile">
                    <div class="update-username">
                        <input v-model="username" placeholder="Enter new username" maxlength="16" minlength="3"/>
                        <button
              class="custom-button"
              @click="updateUsername"
              :disabled="!newUserName || newUserName === userName" >Update Username</button>
          </div>
          <div class="update-photo-section">
            <input type="file" @change="handlePhotoUpload" accept="image/*" />
            <button class="custom-button" @click="updatePhoto" :disabled="!newPhoto">Update Photo</button>Update Photo</button>
          </div>
        </div>
        <ErrorMsg v-if="errormsg" :msg="errormsg" />
      </div>
    </div>
  </div>
</template>
