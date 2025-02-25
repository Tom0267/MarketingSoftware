document.addEventListener('DOMContentLoaded', function () {
    //initialize Quill editor
    var quill = new Quill('#editor', { theme: 'snow' });

    //array for attachment files
    let selectedFiles = [];

    //handle file attachments
    document.getElementById('attachments').addEventListener('change', function (event) {
        for (let file of event.target.files) {
            selectedFiles.push(file);
        }
        updateAttachmentList();
    });

    //toggle schedule options
    const toggleScheduleButton = document.getElementById('toggleScheduleButton');
    const scheduleOptionsContainer = document.getElementById('scheduleOptionsContainer');
    toggleScheduleButton.addEventListener('click', function () {
        scheduleOptionsContainer.classList.toggle('hidden');
    });

    //show/hide custom schedule input
    const scheduleSelect = document.getElementById('schedule');
    const customScheduleContainer = document.getElementById('customScheduleContainer');
    scheduleSelect.addEventListener('change', function () {
        if (this.value === 'custom') {
            customScheduleContainer.classList.remove('hidden');
        } else {
            customScheduleContainer.classList.add('hidden');
        }
    });

    //modal toggling for templates and campaigns
    document.getElementById('openTemplateModal').addEventListener('click', function () {
        loadTemplates();
        document.getElementById('templateModal').classList.remove('hidden');
    });
    document.getElementById('closeTemplateModal').addEventListener('click', function () {
        document.getElementById('templateModal').classList.add('hidden');
    });
    document.getElementById('addTemplateButton').addEventListener('click', function () {
        document.getElementById('addTemplateModal').classList.remove('hidden');
    });
    document.getElementById('closeAddTemplateModal').addEventListener('click', function () {
        document.getElementById('addTemplateModal').classList.add('hidden');
    });
    document.getElementById('openCampaignModal').addEventListener('click', function () {
        document.getElementById('CampaignModal').classList.remove('hidden');
    });
    document.getElementById('closeCampaignModal').addEventListener('click', function () {
        document.getElementById('CampaignModal').classList.add('hidden');
    });
    document.getElementById('cancelCampaign').addEventListener('click', function () {
        document.getElementById('CampaignModal').classList.add('hidden');
    });

    //populate campaign dropdown menu
    const campaignDropdownButton = document.getElementById("campaignDropdownButton");
    const campaignDropdownMenu = document.getElementById("campaignDropdownMenu");
    const selectedCampaign = document.getElementById("selectedCampaign");
    const campaignInput = document.getElementById("campaign");

    //function to fetch campaigns
    async function fetchCampaigns() {
        try {
            const response = await fetch("/campaigns/list"); // Update with your actual API if needed
            if (!response.ok) throw new Error("Failed to fetch campaigns");
    
            const data = await response.json(); // data is an object with a 'campaigns' property
            if (!Array.isArray(data.campaigns)) {
                throw new Error("Invalid campaigns data format");
            }
            populateCampaignDropdown(data.campaigns);
        } catch (error) {
            console.error("Error fetching campaigns:", error);
        }
    }

    //populate campaign dropdown
    function populateCampaignDropdown(campaigns) {
        campaignDropdownMenu.innerHTML = ""; // clear previous entries
    
        campaigns.forEach((campaign) => {
            const option = document.createElement("div");
            option.className = "px-4 py-2 text-gray-700 hover:bg-gray-200 cursor-pointer";
            option.textContent = campaign; // assuming campaign is a string name
    
            option.addEventListener("click", function () {
                document.getElementById("campaignNameDisplay").value = campaign; // update the read-only input
                campaignDropdownMenu.classList.add("hidden");
            });
            campaignDropdownMenu.appendChild(option);
        });
    }

    //toggle dropdown
    campaignDropdownButton.addEventListener("click", () => {
        campaignDropdownMenu.classList.toggle("hidden");
        if (!campaignDropdownMenu.classList.contains("hidden")) {
            fetchCampaigns();
        }
    });

    //close dropdown when clicking outside
    document.addEventListener("click", (event) => {
        if (!campaignDropdownButton.contains(event.target) && !campaignDropdownMenu.contains(event.target)) {
            campaignDropdownMenu.classList.add("hidden");
        }
    });

    //fetch and populate campaigns on load
    fetchCampaigns();

    //email form submission
    document.getElementById('emailForm').addEventListener('submit', function (event) {
        event.preventDefault();
        document.getElementById('body').value = quill.root.innerHTML;
        const sendButton = event.target.querySelector('button[type="submit"]');
        sendButton.disabled = true;
        var formData = new FormData(this);
        //append attachment files
        for (let file of selectedFiles) {
            formData.append('attachments[]', file);
        }
        //validate recipient emails
        var recipients = formData.get('recipients');
        if (!recipients || !validateEmails(recipients.trim())) {
            showNotification('Invalid email format. Use comma-separated valid emails.', 'error');
            sendButton.disabled = false;
            return;
        }
        if (formData.get('subject').trim() === '') {
            showNotification('Please enter a subject', 'error');
            sendButton.disabled = false;
            return;
        }
        if (quill.getText().trim() === '') {
            showNotification('Please enter an email body', 'error');
            sendButton.disabled = false;
            return;
        }
        fetch('/composer', {
            method: 'POST',
            body: formData
        })
            .then(response => {
                if (!response.ok) {
                    return response.json().then(errorData => {
                        showNotification(errorData.message, 'error');
                        throw new Error(errorData.message || "Unknown error");
                    });
                }
                showNotification('Email sent successfully!', 'success');
                quill.setContents([]);
                selectedFiles = [];
                updateAttachmentList();
                document.getElementById('emailForm').reset();
            })
            .catch(error => {
                console.error('Error:', error);
            })
            .finally(() => {
                sendButton.disabled = false;
                selectedFiles = [];
            });
    });

    //email validation function
    function validateEmails(emailString) {
        var emailArray = emailString.split(',').map(email => email.trim());
        var emailRegex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
        return emailArray.every(email => emailRegex.test(email));
    }

    //load canned email templates
    function loadTemplates() {
        fetch('/templates', { method: 'GET' })
            .then(response => response.json())
            .then(templates => {
                if (templates.length === 0) return;
                const templateList = document.getElementById('templateList');
                templateList.innerHTML = '';
                templates.forEach(template => {
                    const li = document.createElement('li');
                    li.classList.add('p-2', 'border', 'rounded', 'cursor-pointer', 'hover:bg-gray-200');
                    li.textContent = template.Title;
                    li.onclick = function () {
                        quill.root.innerHTML = template.Content;
                        document.getElementById('templateModal').classList.add('hidden');
                    };
                    templateList.appendChild(li);
                });
            })
            .catch(error => {
                console.error('Error loading templates:', error);
                showNotification('Error loading templates.', 'error');
            });
    }

    //template form submission
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
                    document.getElementById('addTemplateModal').classList.add('hidden');
                    document.getElementById('templateForm').reset();
                } else {
                    showNotification('Error saving template.', 'error');
                }
            })
            .catch(error => console.error('Error saving template:', error));
    });

    //update attachment list display
    function updateAttachmentList() {
        let fileList = document.getElementById('fileList');
        if (!fileList) {
            fileList = document.createElement('ul');
            fileList.id = 'fileList';
            document.getElementById('attachments').insertAdjacentElement('afterend', fileList);
        }
        fileList.innerHTML = '';
        for (let i = 0; i < selectedFiles.length; i++) {
            let file = selectedFiles[i];
            let li = document.createElement('li');
            li.textContent = file.name + ' ';
            let removeBtn = document.createElement('button');
            removeBtn.textContent = '❌';
            removeBtn.classList.add('ml-2', 'text-red-500', 'hover:text-red-700');
            removeBtn.onclick = function () {
                selectedFiles.splice(i, 1);
                updateAttachmentList();
                document.getElementById('attachments').value = '';
            };
            li.appendChild(removeBtn);
            fileList.appendChild(li);
        }
    }

    document.getElementById('campaignForm').addEventListener('submit', function(event) {
        event.preventDefault();
        const campaignName = document.getElementById('campaignName').value.trim();
        // Convert mailingList from comma-separated string into an array:
        const mailingListValue = document.getElementById('mailingList').value.trim();
        const mailingListArray = mailingListValue
          .split(',')
          .map(email => email.trim())
          .filter(email => email !== "");
        
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
            document.getElementById('CampaignModal').classList.add('hidden');
            document.getElementById('campaignForm').reset();
          } else {
            showNotification('Error creating campaign.', 'error');
          }
        })
        .catch(error => console.error('Error creating campaign:', error));
    });

    //show notification
    function showNotification(message, type) {
        var notification = document.getElementById('notification');
        notification.innerHTML = `<strong>${message}</strong>`;
        notification.classList.remove('hidden');
        notification.classList.remove('bg-green-100', 'text-green-700', 'bg-red-100', 'text-red-700');
        if (type === 'success') {
            notification.classList.add('bg-green-100', 'text-green-700');
        } else if (type === 'error') {
            notification.classList.add('bg-red-100', 'text-red-700');
        }
        setTimeout(function () {
            notification.classList.add('hidden');
        }, 5000);
    }
});
