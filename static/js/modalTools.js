// Tools for modals

// When a user selects a new image for the element with the given inputID, swap the old image for the new one
// in the element with the given imageID
function swap_image(inputID, imageID) {
    let fr = new FileReader();
    fr.readAsDataURL(document.getElementById(inputID).files[0]);
    fr.onload = function (event) {
        document.getElementById(imageID).src = event.target.result;
    };
}

// removes an image from a modal
function remove_image(inputID, imageID, ogSRC) {
    // set image back to original
    document.getElementById(imageID).src = ogSRC;
    document.getElementById(inputID).value = '';
}

// wrapper around the swap_image function for posts that adds a photo to the post and sets "has_image" to true
function add_photo(inputID, imageID, hasImageID) {
    swap_image(inputID, imageID);
    document.getElementById(hasImageID).value = 'true';
}

// wrapper around the remove_image function for posts that removes a photo to the post and sets "has_image" to false
function delete_photo(inputID, imageID, hasImageID, ogSRC) {
    remove_image(inputID, imageID, ogSRC);
    document.getElementById(hasImageID).value = 'false';
}
