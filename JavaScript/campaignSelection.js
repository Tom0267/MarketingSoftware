document.addEventListener('DOMContentLoaded', function () {
    const campaignSearch = document.getElementById("campaignSearch");
    const campaignDropdown = document.getElementById("campaignDropdown");
    const campaignList = document.getElementById("campaignList");
    const selectAllBtn = document.getElementById("selectAllCampaigns");
    const clearAllBtn = document.getElementById("clearAllCampaigns");

    let campaigns = [];

    async function fetchCampaigns() {
        try {
            const response = await fetch("/campaigns/list");
            if (!response.ok) throw new Error("Failed to fetch campaigns");

            const data = await response.json();
            if (!Array.isArray(data.campaigns)) throw new Error("Invalid campaigns format");

            campaigns = data.campaigns;
            populateCampaignList(campaigns);
        } catch (error) {
            console.error("Error fetching campaigns:", error);
        }
    }

    // populate the campaign list with checkboxes
    function populateCampaignList(list) {
        campaignList.innerHTML = ""; // clear old content

        if (list.length === 0) {
            campaignList.innerHTML = "<p class='p-2 text-gray-500'>No campaigns found.</p>";
            return;
        }

        list.forEach(campaign => {
            const label = document.createElement("label");
            label.classList.add("block", "p-2", "cursor-pointer", "hover:bg-gray-100");

            const checkbox = document.createElement("input");
            checkbox.type = "checkbox";
            checkbox.value = campaign;
            checkbox.name = "campaigns";
            checkbox.classList.add("mr-2");

            label.appendChild(checkbox);
            label.appendChild(document.createTextNode(campaign));
            campaignList.appendChild(label);
        });
    }

    // show dropdown when clicking search input
    campaignSearch.addEventListener("focus", () => {
        if (campaigns.length > 0) {
            fetchCampaigns();
            campaignDropdown.classList.remove("hidden");
        }
    });

    // hide dropdown when clicking outside
    document.addEventListener("click", function (event) {
        if (!campaignDropdown.contains(event.target) && event.target !== campaignSearch) {
            campaignDropdown.classList.add("hidden"); 
        }
    });

    // live search filtering
    campaignSearch.addEventListener("input", function () {
        const searchTerm = campaignSearch.value.toLowerCase();
        const filteredCampaigns = searchTerm
            ? campaigns.filter(campaign => campaign.toLowerCase().includes(searchTerm))
            : campaigns;

        populateCampaignList(filteredCampaigns);
    });

    // select all campaigns
    selectAllBtn.addEventListener("click", () => {
        document.querySelectorAll("#campaignList input[type='checkbox']").forEach(checkbox => {
            checkbox.checked = true;
        });
    });

    // clear selections
    clearAllBtn.addEventListener("click", () => {
        document.querySelectorAll("#campaignList input[type='checkbox']").forEach(checkbox => {
            checkbox.checked = false;
        });
    });

    fetchCampaigns();
});