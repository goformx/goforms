<script setup lang="ts">
import { ref, computed } from "vue";
import { useForm, usePage } from "@inertiajs/vue3";
import DashboardLayout from "@/components/layout/DashboardLayout.vue";
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
import { CheckCircle2, AlertCircle } from "lucide-vue-next";

interface User {
  id: string;
  email: string;
  firstName: string;
  lastName: string;
  role: string;
}

interface PageProps {
  auth: {
    user: User;
  };
  flash?: {
    success?: string;
    error?: string;
  };
}

const page = usePage<PageProps>();
const user = computed(() => page.props.auth.user);

const form = useForm({
  firstName: user.value?.firstName ?? "",
  lastName: user.value?.lastName ?? "",
  email: user.value?.email ?? "",
});

const passwordForm = useForm({
  currentPassword: "",
  newPassword: "",
  confirmPassword: "",
});

const successMessage = ref<string | null>(null);
const errorMessage = ref<string | null>(null);

const isSubmitting = computed(() => form.processing);
const isPasswordSubmitting = computed(() => passwordForm.processing);

function handleProfileSubmit() {
  successMessage.value = null;
  errorMessage.value = null;

  form.put("/profile", {
    onSuccess: () => {
      successMessage.value = "Profile updated successfully";
    },
    onError: () => {
      errorMessage.value = "Failed to update profile";
    },
  });
}

function handlePasswordSubmit() {
  successMessage.value = null;
  errorMessage.value = null;

  if (passwordForm.newPassword !== passwordForm.confirmPassword) {
    errorMessage.value = "Passwords don't match";
    return;
  }

  passwordForm.put("/profile/password", {
    onSuccess: () => {
      successMessage.value = "Password updated successfully";
      passwordForm.reset();
    },
    onError: () => {
      errorMessage.value = "Failed to update password";
    },
  });
}
</script>

<template>
  <DashboardLayout title="Profile" subtitle="Manage your account settings">
    <div class="space-y-6">
      <!-- Success/Error Messages -->
      <Alert v-if="successMessage" variant="success">
        <CheckCircle2 class="h-4 w-4" />
        <AlertDescription>{{ successMessage }}</AlertDescription>
      </Alert>

      <Alert v-if="errorMessage" variant="destructive">
        <AlertCircle class="h-4 w-4" />
        <AlertDescription>{{ errorMessage }}</AlertDescription>
      </Alert>

      <!-- Profile Information -->
      <Card class="bg-card/50 backdrop-blur-sm border-border/50">
        <CardHeader>
          <CardTitle>Profile Information</CardTitle>
          <CardDescription>
            Update your account's profile information and email address.
          </CardDescription>
        </CardHeader>
        <form @submit.prevent="handleProfileSubmit">
          <CardContent class="space-y-4">
            <div class="grid grid-cols-2 gap-4">
              <div class="space-y-2">
                <Label for="firstName">First Name</Label>
                <Input id="firstName" v-model="form.firstName" type="text" placeholder="Enter your first name" />
              </div>
              <div class="space-y-2">
                <Label for="lastName">Last Name</Label>
                <Input id="lastName" v-model="form.lastName" type="text" placeholder="Enter your last name" />
              </div>
            </div>
            <div class="space-y-2">
              <Label for="email">Email</Label>
              <Input id="email" v-model="form.email" type="email" placeholder="Enter your email" />
            </div>
          </CardContent>
          <CardFooter>
            <Button type="submit" :disabled="isSubmitting">
              <span v-if="isSubmitting">Saving...</span>
              <span v-else>Save Changes</span>
            </Button>
          </CardFooter>
        </form>
      </Card>

      <!-- Update Password -->
      <Card class="bg-card/50 backdrop-blur-sm border-border/50">
        <CardHeader>
          <CardTitle>Update Password</CardTitle>
          <CardDescription>
            Ensure your account is using a long, random password to stay secure.
          </CardDescription>
        </CardHeader>
        <form @submit.prevent="handlePasswordSubmit">
          <CardContent class="space-y-4">
            <div class="space-y-2">
              <Label for="currentPassword">Current Password</Label>
              <Input id="currentPassword" v-model="passwordForm.currentPassword" type="password"
                placeholder="Enter your current password" />
            </div>
            <div class="space-y-2">
              <Label for="newPassword">New Password</Label>
              <Input id="newPassword" v-model="passwordForm.newPassword" type="password"
                placeholder="Enter your new password" />
            </div>
            <div class="space-y-2">
              <Label for="confirmPassword">Confirm Password</Label>
              <Input id="confirmPassword" v-model="passwordForm.confirmPassword" type="password"
                placeholder="Confirm your new password" />
            </div>
          </CardContent>
          <CardFooter>
            <Button type="submit" :disabled="isPasswordSubmitting">
              <span v-if="isPasswordSubmitting">Updating...</span>
              <span v-else>Update Password</span>
            </Button>
          </CardFooter>
        </form>
      </Card>
    </div>
  </DashboardLayout>
</template>
