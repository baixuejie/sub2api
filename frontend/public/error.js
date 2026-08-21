const modal = document.getElementById('wechat-modal');
const openButton = document.getElementById('wechat-button');
const closeButton = document.getElementById('modal-close');
const emailModal = document.getElementById('email-modal');
const emailButton = document.getElementById('email-button');
const emailCloseButton = document.getElementById('email-modal-close');
const copyEmailButton = document.getElementById('copy-email-button');
const emailAddress = 'baixuejieemail@163.com';

function openModal(target, focusTarget) {
  target.classList.add('is-open');
  document.body.style.overflow = 'hidden';
  focusTarget.focus();
}

function closeModal(target, focusTarget) {
  target.classList.remove('is-open');
  document.body.style.overflow = '';
  focusTarget.focus();
}

openButton.addEventListener('click', () => openModal(modal, closeButton));
closeButton.addEventListener('click', () => closeModal(modal, openButton));
emailButton.addEventListener('click', () => openModal(emailModal, emailCloseButton));
emailCloseButton.addEventListener('click', () => closeModal(emailModal, emailButton));
modal.addEventListener('click', function (event) {
  if (event.target === modal) closeModal(modal, openButton);
});
emailModal.addEventListener('click', function (event) {
  if (event.target === emailModal) closeModal(emailModal, emailButton);
});
copyEmailButton.addEventListener('click', async function () {
  try {
    await navigator.clipboard.writeText(emailAddress);
    copyEmailButton.textContent = '已复制';
    setTimeout(() => { copyEmailButton.textContent = '复制邮箱'; }, 1600);
  } catch {
    copyEmailButton.textContent = emailAddress;
  }
});
document.addEventListener('keydown', function (event) {
  if (event.key !== 'Escape') return;
  if (modal.classList.contains('is-open')) closeModal(modal, openButton);
  if (emailModal.classList.contains('is-open')) closeModal(emailModal, emailButton);
});
