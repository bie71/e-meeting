// snack-management.js

const token = JSON.parse(localStorage.getItem("loginData"))?.token;
if (!token) {
  alert("Silakan login terlebih dahulu.");
  window.location.href = "/api/v1/login";
}

const prevBtn = document.getElementById("prevPageBtn");
const nextBtn = document.getElementById("nextPageBtn");
const currentPageSpan = document.getElementById("currentPage");
let currentPage = 1;
const pageSize = 20;

function updatePaginationButtons(totalPages) {
  prevBtn.disabled = currentPage <= 1;
  nextBtn.disabled = totalPages && currentPage >= totalPages;
  currentPageSpan.textContent = currentPage;
}

async function loadSnacks(filters = {}) {
  const snacksList = document.getElementById("snacksList");
  snacksList.innerHTML = "<p>Loading snacks...</p>";


  try {
    let res;

    let url = `/api/v1/snacks?page=${currentPage}&page_size=${pageSize}`;
    const params = new URLSearchParams();
    
    if (filters.search) params.append("search", filters.search);
    if (filters.category) params.append("category", filters.category);
    if (filters.min_price !== null && filters.min_price !== undefined) params.append("min_price", filters.min_price);
    if (filters.max_price !== null && filters.max_price !== undefined) params.append("max_price", filters.max_price);
    
    const queryString = params.toString();
    if (queryString) {
      url += `&${queryString}`;
    }

    console.log("Request URL:", url);
    
    res = await fetch(url, {
      method: "GET",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });
    

    if (!res.ok) {
        if (res.status === 404) {
          // Misal: tidak ada data ditemukan
          // Bisa langsung set daftar kosong atau tampilkan pesan khusus
          const data = await res.json(); // optional kalau backend kirim pesan error JSON
          throw new Error(data?.error || "Data tidak ditemukan.");
        } else {
          throw new Error("Gagal mengambil daftar snacks.");
        }
      }
      

    const result = await res.json();

    console.log("API Response:", result);

    const snacks = result.snacks || [];
    const totalPages = result.total_pages || null;

    if (snacks.length === 0) {
      snacksList.innerHTML = "<p>Tidak ada snack ditemukan.</p>";
      updatePaginationButtons(totalPages);
      return;
    }

    snacksList.innerHTML = "";
    snacks.forEach((snack) => {
      const snackEl = document.createElement("div");
      snackEl.className = "snack-item";

      snackEl.innerHTML = `
        <div class="snack-info" >
          <h3 class="snack-name">${snack.name}</h3>
          <div class="snack-details">
            <span>Category: ${snack.category}</span>
            <span>Price: $${snack.price.toFixed(2)}</span>
          </div>
        </div>
       <div class="action-buttons">
            <button class="update-btn" title="Update Snack">Update</button>
            <button class="delete-btn" title="Delete Snack">Delete</button>
        </div>
      `;

      snackEl.querySelector(".update-btn").addEventListener("click", () => {
        window.location.href = `snack_detail.html?id=${snack.id}`;
      });

      snackEl.querySelector(".delete-btn").addEventListener("click", async (e) => {
        e.stopPropagation();
        if (!confirm(`Yakin ingin menghapus snack \"${snack.name}\"?`)) return;

        try {
          const delRes = await fetch(`/api/v1/admin/snacks/${snack.id}`, {
            method: "DELETE",
            headers: {
              Authorization: `Bearer ${token}`,
            },
          });

          if (!delRes.ok) {
            const errData = await delRes.json();
            throw new Error(errData.message || "Gagal menghapus snack.");
          }

          alert("Snack berhasil dihapus.");
          loadSnacks(getCurrentFilters());
        } catch (err) {
          alert("Error hapus snack: " + err.message);
        }
      });

      snacksList.appendChild(snackEl);
    });

    updatePaginationButtons(totalPages);
  } catch (err) {
    snacksList.innerHTML = `<p style="color:red;">Error: ${err.message}</p>`;
  }
}

prevBtn.addEventListener("click", () => {
  if (currentPage > 1) {
    currentPage--;
    loadSnacks(getCurrentFilters());
  }
});

nextBtn.addEventListener("click", () => {
  currentPage++;
  loadSnacks(getCurrentFilters());
});

function getCurrentFilters() {
  const form = document.getElementById("searchForm");
  return {
    search: form.search.value.trim(),
    category: form.category.value.trim(),
    min_price: form.min_price.value ? parseFloat(form.min_price.value) : null,
    max_price: form.max_price.value ? parseFloat(form.max_price.value) : null,
  };
}

document.getElementById("searchForm").addEventListener("submit", (e) => {
  e.preventDefault();
  currentPage = 1;
  loadSnacks(getCurrentFilters());
});

document.getElementById("resetFilterBtn").addEventListener("click", () => {
  const filterForm = document.getElementById("searchForm");
  filterForm.reset();
  currentPage = 1;
  loadSnacks();
});

// Load awal
loadSnacks();



document.getElementById('createSnackForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    const alertEl = document.getElementById('createSnackAlert');
    alertEl.style.display = 'none';
  
    const form = e.target;
    const name = form.name.value.trim();
    const category = form.category.value.trim();
    const price = parseFloat(form.price.value);
  
    if (!name || !category || isNaN(price) || price < 0) {
      alertEl.style.color = 'red';
      alertEl.textContent = 'Mohon isi data dengan benar.';
      alertEl.style.display = 'block';
      return;
    }
  
    try {
      const token = JSON.parse(localStorage.getItem('loginData'))?.token;
      if (!token) throw new Error('Token tidak ditemukan, silakan login ulang.');
  
      const payload = { name, category, price };
  
      const res = await fetch('/api/v1/admin/snacks', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify(payload),
      });
  
      if (!res.ok) {
        const errData = await res.json();
        throw new Error(errData.message || 'Gagal membuat snack.');
      }
  
      alertEl.style.color = 'green';
      alertEl.textContent = 'Snack berhasil dibuat!';
      alertEl.style.display = 'block';
      form.reset();
  
      // Optional: reload snack list setelah berhasil create
      loadSnacks();
  
    } catch (err) {
      alertEl.style.color = 'red';
      alertEl.textContent = err.message;
      alertEl.style.display = 'block';
    }
  });
  