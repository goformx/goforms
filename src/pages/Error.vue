<script setup lang="ts">
import { computed } from "vue";
import { Link } from "@inertiajs/vue3";
import GuestLayout from "@/components/layout/GuestLayout.vue";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { AlertTriangle, Home, ArrowLeft } from "lucide-vue-next";

interface Props {
  status?: number;
  message?: string;
}

const props = withDefaults(defineProps<Props>(), {
  status: 500,
  message: "",
});

const title = computed(() => {
  switch (props.status) {
    case 400:
      return "Bad Request";
    case 401:
      return "Unauthorized";
    case 403:
      return "Forbidden";
    case 404:
      return "Page Not Found";
    case 419:
      return "Session Expired";
    case 429:
      return "Too Many Requests";
    case 500:
      return "Server Error";
    case 503:
      return "Service Unavailable";
    default:
      return "Error";
  }
});

const description = computed(() => {
  if (props.message) {
    return props.message;
  }

  switch (props.status) {
    case 400:
      return "The request could not be understood by the server.";
    case 401:
      return "You need to be logged in to access this page.";
    case 403:
      return "You don't have permission to access this resource.";
    case 404:
      return "The page you are looking for doesn't exist or has been moved.";
    case 419:
      return "Your session has expired. Please refresh and try again.";
    case 429:
      return "You've made too many requests. Please wait a moment and try again.";
    case 500:
      return "Something went wrong on our end. Please try again later.";
    case 503:
      return "The service is temporarily unavailable. Please try again later.";
    default:
      return "An unexpected error occurred.";
  }
});

function goBack() {
  window.history.back();
}
</script>

<template>
  <GuestLayout :title="title">
    <div class="relative flex min-h-[calc(100vh-8rem)] items-center justify-center px-4 py-12">
      <!-- Subtle gradient background -->
      <div class="absolute inset-0 overflow-hidden">
        <div class="absolute top-[30%] left-[20%] w-[400px] h-[400px] bg-red-500/10 rounded-full blur-3xl" />
        <div class="absolute bottom-[20%] right-[20%] w-[300px] h-[300px] bg-orange-500/10 rounded-full blur-3xl" />
      </div>

      <Card class="relative z-10 w-full max-w-md text-center bg-card/80 backdrop-blur-sm border-border/50">
        <CardHeader>
          <div class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-destructive/10">
            <AlertTriangle class="h-8 w-8 text-destructive" />
          </div>
          <CardTitle class="text-4xl font-bold">{{ status }}</CardTitle>
          <h2 class="text-xl font-semibold text-muted-foreground">{{ title }}</h2>
        </CardHeader>

        <CardContent>
          <p class="text-muted-foreground">{{ description }}</p>
        </CardContent>

        <CardFooter class="flex flex-col gap-2 sm:flex-row sm:justify-center">
          <Button variant="outline" @click="goBack">
            <ArrowLeft class="mr-2 h-4 w-4" />
            Go Back
          </Button>
          <Button as-child>
            <Link href="/">
              <Home class="mr-2 h-4 w-4" />
              Go Home
            </Link>
          </Button>
        </CardFooter>
      </Card>
    </div>
  </GuestLayout>
</template>
