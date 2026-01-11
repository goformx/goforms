<script setup lang="ts">
import { onMounted, onUnmounted } from "vue";
import { Head } from "@inertiajs/vue3";
import Nav from "@/components/shared/Nav.vue";
import Footer from "@/components/shared/Footer.vue";

interface Props {
  title?: string;
}

withDefaults(defineProps<Props>(), {
  title: "Dashboard",
});

// Apply dark mode to html element for Tailwind
onMounted(() => {
  document.documentElement.classList.add("dark");
});

onUnmounted(() => {
  document.documentElement.classList.remove("dark");
});
</script>

<template>
  <div class="min-h-screen flex flex-col bg-background">

    <Head :title="title" />

    <header>
      <Nav />
    </header>

    <main class="flex-1 container py-6">
      <slot />
    </main>

    <Footer />
  </div>
</template>
