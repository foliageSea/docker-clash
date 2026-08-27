import { createApp } from 'vue'
import App from './App.vue'
import 'vue-sonner/style.css'
import './style.css'

const savedTheme = localStorage.getItem('theme')
const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
document.documentElement.classList.toggle('dark', savedTheme ? savedTheme === 'dark' : prefersDark)

createApp(App).mount('#app')
