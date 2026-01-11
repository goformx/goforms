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
import { useFormValidation, signupSchema, type SignupFormData } from "@/composables/useFormValidation";
import { AlertCircle } from "lucide-vue-next";

interface Props {
  flash?: {
    success?: string;
    error?: string;
  };
}

const props = defineProps<Props>();

const { errors: validationErrors, validate, validateField, clearFieldError } = useFormValidation(signupSchema);

const form = useForm<SignupFormData>({
  email: "",
  password: "",
  confirmPassword: "",
});

const serverError = ref<string | null>(null);

const isSubmitting = computed(() => form.processing);

function handleFieldBlur(field: keyof SignupFormData) {
  validateField(field, form[field]);
}

function handleFieldInput(field: keyof SignupFormData) {
  clearFieldError(field);
  serverError.value = null;
}

async function handleSubmit() {
  serverError.value = null;

  const result = validate({
    email: form.email,
    password: form.password,
    confirmPassword: form.confirmPassword,
  });

  if (!result.valid) {
    return;
  }

  form.post("/signup", {
    onError: (errors) => {
      if (errors.email) {
        serverError.value = errors.email;
      } else if (errors.password) {
        serverError.value = errors.password;
      } else {
        serverError.value = "Failed to create account. Please try again.";
      }
    },
  });
}
</script>

<template>
  <GuestLayout title="Sign Up">
    <div class="flex min-h-[calc(100vh-8rem)] items-center justify-center px-4 py-12">
      <Card class="w-full max-w-md">
        <CardHeader class="space-y-1">
          <CardTitle class="text-2xl font-bold text-center">
            Create an account
          </CardTitle>
          <CardDescription class="text-center">
            Enter your details to create your account
          </CardDescription>
        </CardHeader>

        <form @submit.prevent="handleSubmit">
          <CardContent class="space-y-4">
            <!-- Server Error Alert -->
            <Alert v-if="serverError || props.flash?.error" variant="destructive">
              <AlertCircle class="h-4 w-4" />
              <AlertDescription>
                {{ serverError || props.flash?.error }}
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
                placeholder="Create a password"
                autocomplete="new-password"
                :class="{ 'border-destructive': validationErrors.password }"
                @blur="handleFieldBlur('password')"
                @input="handleFieldInput('password')"
              />
              <p v-if="validationErrors.password" class="text-sm text-destructive">
                {{ validationErrors.password }}
              </p>
              <p class="text-xs text-muted-foreground">
                Must be at least 8 characters with uppercase, lowercase, number, and special character
              </p>
            </div>

            <!-- Confirm Password Field -->
            <div class="space-y-2">
              <Label for="confirmPassword">Confirm Password</Label>
              <Input
                id="confirmPassword"
                v-model="form.confirmPassword"
                type="password"
                placeholder="Confirm your password"
                autocomplete="new-password"
                :class="{ 'border-destructive': validationErrors.confirmPassword }"
                @blur="handleFieldBlur('confirmPassword')"
                @input="handleFieldInput('confirmPassword')"
              />
              <p v-if="validationErrors.confirmPassword" class="text-sm text-destructive">
                {{ validationErrors.confirmPassword }}
              </p>
            </div>
          </CardContent>

          <CardFooter class="flex flex-col space-y-4">
            <Button
              type="submit"
              class="w-full"
              :disabled="isSubmitting"
            >
              <span v-if="isSubmitting">Creating account...</span>
              <span v-else>Create account</span>
            </Button>

            <div class="text-center text-sm text-muted-foreground">
              Already have an account?
              <a
                href="/login"
                class="text-primary hover:underline"
                @click.prevent="router.visit('/login')"
              >
                Sign in
              </a>
            </div>
          </CardFooter>
        </form>
      </Card>
    </div>
  </GuestLayout>
</template>
