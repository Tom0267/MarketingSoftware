import { showNotification } from './utils.js';
document.getElementById('campaignForm').addEventListener('submit', function (event) {
    event.preventDefault();

    const campaignName = document.getElementById('campaignName').value.trim();
    const mailingListValue = document.getElementById('mailingList').value.trim();
    
    // convert mailing list to array
    const mailingListArray = mailingListValue.split(',')
        .map(email => email.trim())
        .filter(email => email !== ""); // remove empty values

    // validate input
    if (!campaignName || mailingListArray.length === 0) {
        showNotification('Please enter a campaign name and valid mailing list.', 'error');
        return;
    }

    fetch('/campaigns', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ campaignName, mailingList: mailingListArray })
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            showNotification('Campaign created successfully!', 'success');
            document.getElementById('CampaignModal').classList.add('hidden'); // close modal
            document.getElementById('campaignForm').reset(); // reset form
        } else {
            showNotification('Error creating campaign.', 'error');
        }
    })
    .catch(error => console.error('Error creating campaign:', error));
});