<template>
  <div class="landing-page">
    <div class="container">
      <header class="header">
        <h1 class="logo">Spotiarch</h1>
      </header>

      <main class="main-content">
        <section class="hero">
          <h2 class="hero-title">Welcome to Spotiarch</h2>
          <p class="hero-subtitle">
            Your personal music archive and discovery platform.
            Organize, explore, and enjoy your music collection like never before.
          </p>
          <div class="cta-buttons">
            <button class="cta-button" @click="goToSignup">
              Get Started
            </button>
            <button class="cta-button cta-login" @click="goToLogin">
              Login
            </button>
          </div>
        </section>

        <section class="features">
          <div class="feature-card">
            <div class="feature-icon">🎵</div>
            <h3>Organize Your Music</h3>
            <p>Keep your entire music collection organized in one place with smart playlists and tags.</p>
          </div>
          <div class="feature-card">
            <div class="feature-icon">🔍</div>
            <h3>Discover New Tracks</h3>
            <p>Find new music based on your listening habits and preferences with intelligent recommendations.</p>
          </div>
          <div class="feature-card">
            <div class="feature-icon">📊</div>
            <h3>Track Your Stats</h3>
            <p>View detailed analytics about your listening patterns and favorite artists.</p>
          </div>
        </section>
      </main>

      <footer class="footer">
        <p>&copy; 2026 Spotiarch. All rights reserved.</p>
      </footer>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()

function goToSignup() {
  router.push({ name: 'signup' })
}

function goToLogin() {
  router.push({ name: 'login' })
}
</script>

<style scoped>
.landing-page {
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

.header {
  padding: 20px 0;
}

.logo {
  font-size: 32px;
  font-weight: bold;
}

.main-content {
  padding: 60px 0;
}

.hero {
  text-align: center;
  margin-bottom: 80px;
}

.hero-title {
  font-size: 56px;
  font-weight: bold;
  margin-bottom: 20px;
}

.hero-subtitle {
  font-size: 20px;
  line-height: 1.6;
  max-width: 600px;
  margin: 0 auto 40px;
  opacity: 0.9;
}

.cta-buttons {
  display: flex;
  gap: 16px;
  justify-content: center;
}

.cta-button {
  background-color: white;
  color: #667eea;
  font-size: 18px;
  font-weight: 600;
  padding: 16px 48px;
  border-radius: 30px;
  box-shadow: 0 4px 15px rgba(0, 0, 0, 0.2);
}

.cta-login {
  background-color: transparent;
  color: white;
  border: 2px solid white;
  box-shadow: none;
}

.features {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 30px;
  margin-top: 60px;
}

.feature-card {
  background-color: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  padding: 40px 30px;
  border-radius: 12px;
  text-align: center;
  transition: transform 0.3s ease;
}

.feature-card:hover {
  transform: translateY(-5px);
  background-color: rgba(255, 255, 255, 0.15);
}

.feature-icon {
  font-size: 48px;
  margin-bottom: 20px;
}

.feature-card h3 {
  font-size: 24px;
  margin-bottom: 12px;
}

.feature-card p {
  font-size: 16px;
  line-height: 1.5;
  opacity: 0.9;
}

.footer {
  text-align: center;
  padding: 40px 0;
  margin-top: 80px;
  border-top: 1px solid rgba(255, 255, 255, 0.2);
  opacity: 0.8;
}
</style>
