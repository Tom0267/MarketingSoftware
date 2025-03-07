export function showNotification(message, type) {
    const notification = document.getElementById('notification');
    notification.innerHTML = `<strong>${message}</strong>`;
    notification.className = `p-4 mt-4 rounded-md ${type === 'success' ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'}`;
    setTimeout(() => notification.classList.add('hidden'), 5000);
}