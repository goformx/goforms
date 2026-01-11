<script setup lang="ts">
import { ref, computed } from "vue";
import { useForm, router } from "@inertiajs/vue3";
import GuestLayout from "@/components/layout/GuestLayout.vue";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { useFormValidation, loginSchema, type LoginFormData } from "@/composables/useFormValidation";
import { AlertCircle } from "lucide-vue-next";

interface Props {
  isDevelopment?: boolean;
  flash?: {
    success?: string;
    error?: string;
  };
}

const props = withDefaults(defineProps<Props>(), {
  isDevelopment: false,
});

const { errors: validationErrors, validate, validateField, clearFieldError } = useFormValidation(loginSchema);

const form = useForm<LoginFormData>({
  email: props.isDevelopment ? "test@example.com" : "",
  password: props.isDevelopment ? "Test123!" : "",
});

const serverError = ref<string | null>(null);

const isSubmitting = computed(() => form.processing);

function handleFieldBlur(field: keyof LoginFormData) {
  validateField(field, form[field]);
}

function handleFieldInput(field: keyof LoginFormData) {
  clearFieldError(field);
  serverError.value = null;
}

async function handleSubmit() {
  serverError.value = null;

  const result = validate({
    email: form.email,
    password: form.password,
  });

  if (!result.valid) {
    return;
  }

  form.post("/login", {
    onError: (errors) => {
      if (errors.email) {
        serverError.value = errors.email;
      } else if (errors.password) {
        serverError.value = errors.password;
      } else {
        serverError.value = "Invalid email or password";
      }
    },
  });
}
</script>

<template>
  <GuestLayout title="Login">
    <div class="flex min-h-[calc(100vh-8rem)] items-center justify-center px-4 py-12">
      <Card class="w-full max-w-md">
        <CardHeader class="space-y-1">
          <CardTitle class="text-2xl font-bold text-center">
            Sign in to your account
          </CardTitle>
          <CardDescription class="text-center">
            Enter your email and password to access your dashboard
          </CardDescription>
        </CardHeader>

        <form @submit.prevent="handleSubmit">
          <CardContent class="space-y-4">
            <!-- Server Error Alert -->
            <Alert v-if="serverError || flash?.error" variant="destructive">
              <AlertCircle class="h-4 w-4" />
              <AlertDescription>
                {{ serverError || flash?.error }}
              </AlertDescription>
            </Alert>

            <!-- Success Alert -->
            <Alert v-if="flash?.success" variant="success">
              <AlertDescription>
                {{ flash.success }}
              </AlertDescription>
            </Alert>

            <!-- Email Field -->
            <div class="space-y-2">
              <Label for="email">Email</Label>
              <Input
                id="email"
                v-model="form.email"
                type="email"
                placeholder="Enter your email"
                autocomplete="email"
                :class="{ 'border-destructive': validationErrors.email }"
                @blur="handleFieldBlur('email')"
                @input="handleFieldInput('email')"
              />
              <p v-if="validationErrors.email" class="text-sm text-destructive">
                {{ validationErrors.email }}
              </p>
            </div>

            <!-- Password Field -->
            <div class="space-y-2">
              <Label for="password">Password</Label>
              <Input
                id="password"
                v-model="form.password"
                type="password"
                placeholder="Enter your password"
                autocomplete="current-password"
                :class="{ 'border-destructive': validationErrors.password }"
                @blur="handleFieldBlur('password')"
                @input="handleFieldInput('password')"
              />
              <p v-if="validationErrors.password" class="text-sm text-destructive">
                {{ validationErrors.password }}
              </p>
            </div>
          </CardContent>

          <CardFooter class="flex flex-col space-y-4">
            <Button
              type="submit"
              class="w-full"
              :disabled="isSubmitting"
            >
              <span v-if="isSubmitting">Signing in...</span>
              <span v-else>Sign in</span>
            </Button>

            <div class="text-center text-sm">
              <a
                href="/forgot-password"
                class="text-primary hover:underline"
                @click.prevent="router.visit('/forgot-password')"
              >
                Forgot your password?
              </a>
            </div>

            <div class="text-center text-sm text-muted-foreground">
              Don't have an account?
              <a
                href="/signup"
                class="text-primary hover:underline"
                @click.prevent="router.visit('/signup')"
              >
                Sign up
              </a>
            </div>
          </CardFooter>
        </form>
      </Card>
    </div>
  </GuestLayout>
</template>
