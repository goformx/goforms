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
import { useFormValidation, forgotPasswordSchema, type ForgotPasswordFormData } from "@/composables/useFormValidation";
import { AlertCircle, CheckCircle2, ArrowLeft } from "lucide-vue-next";

interface Props {
  flash?: {
    success?: string;
    error?: string;
  };
}

const props = defineProps<Props>();

const { errors: validationErrors, validate, validateField, clearFieldError } = useFormValidation(forgotPasswordSchema);

const form = useForm<ForgotPasswordFormData>({
  email: "",
});

const serverError = ref<string | null>(null);
const emailSent = ref(false);

const isSubmitting = computed(() => form.processing);

function handleFieldBlur(field: keyof ForgotPasswordFormData) {
  validateField(field, form[field]);
}

function handleFieldInput(field: keyof ForgotPasswordFormData) {
  clearFieldError(field);
  serverError.value = null;
}

async function handleSubmit() {
  serverError.value = null;

  const result = validate({
    email: form.email,
  });

  if (!result.valid) {
    return;
  }

  form.post("/forgot-password", {
    onSuccess: () => {
      emailSent.value = true;
    },
    onError: (errors) => {
      if (errors.email) {
        serverError.value = errors.email;
      } else {
        serverError.value = "Failed to send reset email. Please try again.";
      }
    },
  });
}
</script>

<template>
  <GuestLayout title="Forgot Password">
    <div class="relative flex min-h-[calc(100vh-8rem)] items-center justify-center px-4 py-12">
      <!-- Subtle gradient background -->
      <div class="absolute inset-0 overflow-hidden">
        <div class="absolute top-[30%] left-[20%] w-[400px] h-[400px] bg-indigo-500/10 rounded-full blur-3xl" />
        <div class="absolute bottom-[20%] right-[20%] w-[300px] h-[300px] bg-purple-500/10 rounded-full blur-3xl" />
      </div>
      
      <Card class="relative z-10 w-full max-w-md bg-card/80 backdrop-blur-sm border-border/50">
        <CardHeader class="space-y-1">
          <CardTitle class="text-2xl font-bold text-center">
            Reset your password
          </CardTitle>
          <CardDescription class="text-center">
            Enter your email address and we'll send you a link to reset your password
          </CardDescription>
        </CardHeader>

        <template v-if="emailSent">
          <CardContent class="space-y-4">
            <Alert variant="success">
              <CheckCircle2 class="h-4 w-4" />
              <AlertDescription>
                If an account exists with that email, you will receive a password reset link shortly.
              </AlertDescription>
            </Alert>
          </CardContent>

          <CardFooter>
            <Button
              variant="outline"
              class="w-full"
              @click="router.visit('/login')"
            >
              <ArrowLeft class="mr-2 h-4 w-4" />
              Back to login
            </Button>
          </CardFooter>
        </template>

        <form v-else @submit.prevent="handleSubmit">
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
          </CardContent>

          <CardFooter class="flex flex-col space-y-4">
            <Button
              type="submit"
              class="w-full"
              :disabled="isSubmitting"
            >
              <span v-if="isSubmitting">Sending...</span>
              <span v-else>Send reset link</span>
            </Button>

            <div class="text-center text-sm text-muted-foreground">
              Remember your password?
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
