import { ref, onMounted } from 'vue';
import { System } from '@wailsio/runtime';

const isMacOS = ref(false);
const isWindows = ref(false);
const isLinux = ref(false);
const platformDetected = ref(false);

export function usePlatform() {
  onMounted(async () => {
    if (platformDetected.value) {
      return; // Already detected
    }

    try {
      const env = await System.Environment();
      isMacOS.value = env.OS === 'darwin';
      isWindows.value = env.OS === 'windows';
      isLinux.value = env.OS === 'linux';
      platformDetected.value = true;
    } catch (error) {
      console.error('Failed to detect platform:', error);
      // Fallback to user agent detection
      const ua = navigator.userAgent.toLowerCase();
      isMacOS.value = ua.includes('mac');
      isWindows.value = ua.includes('win');
      isLinux.value = ua.includes('linux');
      platformDetected.value = true;
    }
  });

  return {
    isMacOS,
    isWindows,
    isLinux,
    platformDetected,
  };
}
