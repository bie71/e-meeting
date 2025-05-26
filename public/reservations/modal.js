const token = JSON.parse(localStorage.getItem('loginData') || '{}').token;
function showDetailInModal(reservationId) {
    axios.get(`/api/v1/reservation/${reservationId}`, {
      headers: { Authorization: `Bearer ${token}` }
    }).then(res => {
      const d = res.data;
      const toRupiah = val => "Rp " + val.toLocaleString('id-ID');
  

    //   pending: 'bg-yellow-100 text-yellow-800',
    //   confirmed: 'bg-green-100 text-green-800',
    //   cancelled: 'bg-red-100 text-red-800',
    //   completed: 'bg-blue-100 text-blue-800'

    const statusClassMap = {
        pending: 'bg-yellow-100 text-yellow-800',
        confirmed: 'bg-green-100 text-green-800',
        cancelled: 'bg-red-100 text-red-800',
        completed: 'bg-blue-100 text-blue-800'
      };
      
      const statusClass = statusClassMap[d.status] || 'bg-gray-100 text-gray-800';
      

        document.getElementById('modalContent').innerHTML = `
        <div>
            <strong>Status:</strong>
            <span class="inline-block px-2 py-1 rounded font-semibold text-xs ${statusClass}">
            ${d.status.toUpperCase()}
            </span>
        </div>
        <div><strong>Mulai:</strong> ${pretty(d.start_time)}</div>
        <div><strong>Selesai:</strong> ${pretty(d.end_time)}</div>
        <div><strong>Pengunjung:</strong> ${d.visitor_count}</div>
        <div><strong>Total:</strong> ${toRupiah(d.total_cost)}</div>
        <hr class="my-2 border-gray-300">
        <div><strong>Ruangan:</strong> ${d.room.name} ( Kapasitas ${d.room.capacity} org, ${toRupiah(d.room.price_per_hour)}/jam)</div>
        <div><strong>User:</strong> ${d.user.username}</div>
        <hr class="my-2 border-gray-300">
      
        <h3 class="mt-4 font-semibold text-blue-700">Daftar Snack</h3>
        <div class="overflow-x-auto mt-2 rounded-lg shadow-sm">
          <table class="min-w-full text-sm text-left border border-gray-300">
            <thead class="bg-blue-600 text-white">
              <tr>
                <th class="px-4 py-2">Nama</th>
                <th class="px-4 py-2">Kategori</th>
                <th class="px-4 py-2">Jumlah</th>
                <th class="px-4 py-2">Harga</th>
                <th class="px-4 py-2">Subtotal</th>
              </tr>
            </thead>
            <tbody>
              ${
                d.snacks.length > 0
                  ? d.snacks.map((s, i) => `
                    <tr class="${i % 2 === 0 ? 'bg-white' : 'bg-blue-50'}">
                      <td class="px-4 py-2 hover:bg-blue-100">${s.name}</td>
                      <td class="px-4 py-2 hover:bg-blue-100">${s.category}</td>
                      <td class="px-4 py-2 hover:bg-blue-100">${s.quantity}</td>
                      <td class="px-4 py-2 hover:bg-blue-100">${toRupiah(s.price)}</td>
                      <td class="px-4 py-2 hover:bg-blue-100">${toRupiah(s.subtotal)}</td>
                    </tr>
                  `).join('')
                  : '<tr><td colspan="5" class="px-4 py-2 text-center text-gray-500">Tidak ada snack</td></tr>'
              }
            </tbody>
          </table>
        </div>
      `;


      if (getUserRole() === 'admin') {
        document.getElementById('modalContent').innerHTML += `
          <div class="mt-4 border-t pt-4 flex flex-col sm:flex-row justify-between">
            <div class="flex flex-col sm:flex-row items-center gap-3">
              <select id="statusSelect" class="border rounded px-3 py-2 w-full sm:w-auto">
                <option value="pending">Pending</option>
                <option value="confirmed">Confirmed</option>
                <option value="cancelled">Cancelled</option>
                <option value="completed">Completed</option>
              </select>
              <button onclick="updateReservationStatus('${d.id}')" class="bg-green-600 hover:bg-green-800 text-white font-semibold px-4 py-2 rounded shadow">
                Update Status
              </button>
              </div>
              <botton 
              onclick="deleteReservation('${d.id}')" class="bg-red-600 hover:bg-red-800 text-white font-semibold px-4 py-2 rounded shadow ">
              Hapus Reservasi
              </botton>
          </div>
        `;
      }
      
      document.getElementById('modalDetail').classList.remove('hidden');
    }).catch(err => {
      alert('Gagal memuat detail reservasi: ' + err.message);
    });
  }
  
  function closeModal() {
    document.getElementById('modalDetail').classList.add('hidden');
  }
  

  const pretty = str => {
    if (!str.includes('T')) return str;
    const [date, time] = str.split('T');
    return `${date} ${time.replace(/:\d\d(\.\d+)?(Z|[+-]\d\d:\d\d)?$/, '')}`;
  };



  function updateReservationStatus(reservationId) {
    
    const selectedStatus = document.getElementById('statusSelect').value;
  
    if (!selectedStatus) {
      alert('Pilih status terlebih dahulu.');
      return;
    }
  
    if (!confirm(`Yakin ingin mengubah status menjadi "${selectedStatus}"?`)) return;
  
    axios.post('/api/v1/admin/reservation/status', {
      reservation_id: reservationId,
      status: selectedStatus
    }, {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    })
    .then(() => {
      alert(`Status berhasil diperbarui ke "${selectedStatus.toUpperCase()}"`);
      closeModal();

      setTimeout(() => {
        window.location.reload();
      }, 1000)
    })
    .catch(err => {
      alert('Gagal update status: ' + (err.response?.data?.message || err.message));
    });
  }
  
  function getUserRole() {
    const token = JSON.parse(localStorage.getItem('loginData') || '{}').token;
    if (!token) return null;
  
    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      return payload.role;
    } catch (e) {
      console.error('Failed to parse JWT:', e);
      return null;
    }
  }
  
  function deleteReservation(reservationId) {
    if (!confirm('Yakin ingin menghapus reservasi ini?')) return;
  
    axios.delete(`/api/v1/admin/reservation/${reservationId}`, {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    })
    .then(res => {
      console.log('✅ DELETE response:', res);
      alert('Reservasi berhasil dihapus');
      closeModal();
      window.location.reload();
    })
    .catch(err => {
      console.error('❌ DELETE error:', err);
      if (err.response) {
        console.error('❌ Server responded with:', err.response.data);
      }
      alert('Gagal menghapus reservasi: ' + (err.response?.data?.error || err.message));
    });
  }
  