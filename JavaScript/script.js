import { showNotification } from './utils.js';
document.addEventListener('DOMContentLoaded', function () {
    // initialize Quill editor
    var quill = new Quill('#editor', {
        theme: 'snow',
        modules: {
            toolbar: [
                [{ 'header': [1, 2, 3, false] }],
                ['bold', 'italic', 'underline', 'strike'],
                [{ 'color': [] }, { 'background': [] }],
                [{ 'script': 'sub' }, { 'script': 'super' }],
                [{ 'list': 'ordered' }, { 'list': 'bullet' }],
                [{ 'align': [] }],
                ['blockquote', 'code-block'],
                ['link', 'image', 'video'],
                ['clean']
            ]
        }
    });

    // handle file attachments
    let selectedFiles = [];
    document.getElementById('attachments').addEventListener('change', function (event) {
        for (let file of event.target.files) {
            selectedFiles.push(file);
        }
        updateAttachmentList();
    });

    // toggle schedule options
    document.getElementById('toggleScheduleButton').addEventListener('click', function () {
        document.getElementById('scheduleOptionsContainer').classList.toggle('hidden');
    });

    // show/hide custom schedule input
    document.getElementById('schedule').addEventListener('change', function () {
        const customScheduleContainer = document.getElementById('customScheduleContainer');
        this.value === 'custom'
            ? customScheduleContainer.classList.remove('hidden')
            : customScheduleContainer.classList.add('hidden');
    });

    // modal controls for templates/campaigns
    function toggleModal(modalId, show) {
        document.getElementById(modalId).classList.toggle('hidden', !show);
    }

    document.getElementById('openTemplateModal').addEventListener('click', function () {
        loadTemplates(); // ensure templates are loaded on modal open
        document.getElementById('templateModal').classList.remove('hidden');
    });
    document.getElementById('closeTemplateModal').addEventListener('click', () => toggleModal('templateModal', false));
    document.getElementById('addTemplateButton').addEventListener('click', () => toggleModal('addTemplateModal', true));
    document.getElementById('closeAddTemplateModal').addEventListener('click', () => toggleModal('addTemplateModal', false));
    document.getElementById('openCampaignModal').addEventListener('click', () => toggleModal('CampaignModal', true));
    document.getElementById('closeCampaignModal').addEventListener('click', () => toggleModal('CampaignModal', false));
    document.getElementById('cancelCampaign').addEventListener('click', () => toggleModal('CampaignModal', false));

    // email form submission
    document.getElementById('emailForm').addEventListener('submit', function (event) {
        event.preventDefault();
        document.getElementById('body').value = quill.root.innerHTML;
        
        const sendButton = event.target.querySelector('button[type="submit"]');
        sendButton.disabled = true;
        let formData = new FormData(this);

        // append selected campaign names
        const selectedCampaigns = getSelectedCampaigns();
        if (selectedCampaigns.length > 0) {
            formData.append('campaigns', selectedCampaigns.join(","));
        }

        if (!validateEmailForm(formData, selectedCampaigns)) {
            sendButton.disabled = false;
            return;
        }

        fetch('/composer', {
            method: 'POST',
            body: formData
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                showNotification('Email sent successfully!', 'success');
                quill.setContents([]);
                selectedFiles = [];
                updateAttachmentList();
                document.getElementById('emailForm').reset();
            } else {
                showNotification('Error sending email.', 'error');
            }
        })
        .catch(error => console.error('Error:', error))
        .finally(() => sendButton.disabled = false);
    });

    function getSelectedCampaigns() {
        return Array.from(document.querySelectorAll("#campaignList input[type='checkbox']:checked"))
            .map(checkbox => checkbox.value);
    }

    function validateEmailForm(formData, selectedCampaigns) {
        let recipients = formData.get('recipients').trim();

        if (!recipients && selectedCampaigns.length === 0) {
            showNotification('Please enter at least one recipient or select a campaign.', 'error');
            return false;
        }

        if (recipients && !validateEmails(recipients)) {
            showNotification('Invalid email format. Use comma-separated emails.', 'error');
            return false;
        }

        if (!formData.get('subject') || formData.get('subject').trim() === '') {
            showNotification('Please enter a subject.', 'error');
            return false;
        }

        if (quill.getText().trim() === '') {
            showNotification('Please enter an email body.', 'error');
            return false;
        }

        return true;
    }

    // email validation function
    function validateEmails(emailString) {
        const emailRegex = /^[a-zA-Z0-9._%+-]+@([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$/;
        return emailString.split(',').map(email => email.trim()).every(email => emailRegex.test(email));
    }

    // load canned email templates
    function loadTemplates() {
        fetch('/templates')
            .then(response => response.json())
            .then(data => {
                if (!Array.isArray(data.templates)) {
                    console.error("Error: Invalid templates format");
                    return;
                }

                const templateList = document.getElementById('templateList');
                templateList.innerHTML = '';

                data.templates.forEach(template => {
                    const li = document.createElement('li');
                    li.classList.add('p-2', 'border', 'rounded', 'cursor-pointer', 'hover:bg-gray-200');
                    li.textContent = template.Title || "Untitled";
                    li.onclick = () => {
                        quill.root.innerHTML = template.Content;
                        toggleModal('templateModal', false);
                    };
                    templateList.appendChild(li);
                });
            })
            .catch(error => console.error("Error loading templates:", error));
    }

    // template form submission
    document.getElementById('templateForm').addEventListener('submit', function (event) {
        event.preventDefault();
        const templateName = document.getElementById('templateName').value.trim();
        const templateContent = document.getElementById('templateContent').value.trim();

        if (!templateName || !templateContent) {
            showNotification('Please enter a template name and content.', 'error');
            return;
        }

        fetch('/templates', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ Title: templateName, Content: templateContent })
        })
        .then(response => response.json())
        .then(data => {
            if (data.success === "true") {
                showNotification('Template saved successfully!', 'success');
                loadTemplates();
                toggleModal('addTemplateModal', false);
                document.getElementById('templateForm').reset();
            } else {
                showNotification('Error saving template.', 'error');
            }
        })
        .catch(error => console.error('Error saving template:', error));
    });

    // update attachment list
    function updateAttachmentList() {
        let fileList = document.getElementById('fileList');
        fileList.innerHTML = selectedFiles.map((file, i) => `
            <li>
                ${file.name} 
                <button class="ml-2 text-red-500 hover:text-red-700" onclick="removeAttachment(${i})">❌</button>
            </li>
        `).join('');
    }

    // remove attachment
    window.removeAttachment = function (index) {
        selectedFiles.splice(index, 1);
        updateAttachmentList();
    };
});