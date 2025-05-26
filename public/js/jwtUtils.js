// jwtUtils.js
function parseJwt(token) {
    try {
      const base64Url = token.split('.')[1];
      const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
      const jsonPayload = decodeURIComponent(atob(base64).split('').map(c =>
        '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2)
      ).join(''));
      return JSON.parse(jsonPayload);
    } catch (e) {
      console.error('Failed to parse JWT:', e);
      return null;
    }
  }
  
  // Buat fungsi ini jadi global supaya bisa dipakai di script lain
  window.parseJwt = parseJwt;
  