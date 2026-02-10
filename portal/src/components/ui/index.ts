/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * VirtEngine Design System Components
 *
 * This module exports all UI components built with Radix UI primitives
 * and styled with Tailwind CSS. All components follow WCAG AA accessibility
 * guidelines and support dark mode.
 */

// Button
export { Button, buttonVariants } from './Button';
export type { ButtonProps } from './Button';

// Card
export { Card, CardHeader, CardFooter, CardTitle, CardDescription, CardContent } from './Card';

// Input
export { Input } from './Input';
export type { InputProps } from './Input';

// Label
export { Label } from './Label';

// Select
export {
  Select,
  SelectGroup,
  SelectValue,
  SelectTrigger,
  SelectContent,
  SelectLabel,
  SelectItem,
  SelectSeparator,
  SelectScrollUpButton,
  SelectScrollDownButton,
} from './Select';

// Modal / Dialog
export {
  Dialog,
  DialogPortal,
  DialogOverlay,
  DialogClose,
  DialogTrigger,
  DialogContent,
  DialogHeader,
  DialogFooter,
  DialogTitle,
  DialogDescription,
  Modal,
} from './Modal';

// Toast
export {
  ToastProvider,
  ToastViewport,
  Toast,
  ToastTitle,
  ToastDescription,
  ToastClose,
  ToastAction,
} from './Toast';
export type { ToastProps, ToastActionElement } from './Toast';
export { Toaster } from './Toaster';

// Tabs
export { Tabs, TabsList, TabsTrigger, TabsContent } from './Tabs';

// Table
export {
  Table,
  TableHeader,
  TableBody,
  TableFooter,
  TableHead,
  TableRow,
  TableCell,
  TableCaption,
} from './Table';

// Badge
export { Badge, badgeVariants } from './Badge';
export type { BadgeProps } from './Badge';

// Avatar
export { Avatar, AvatarImage, AvatarFallback, AvatarGroup } from './Avatar';

// Skeleton
export { Skeleton, SkeletonText, SkeletonCard, SkeletonAvatar, SkeletonTable } from './Skeleton';
export type { SkeletonProps } from './Skeleton';

// Tooltip
export { Tooltip, TooltipTrigger, TooltipContent, TooltipProvider, SimpleTooltip } from './Tooltip';

// Dropdown Menu
export {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuCheckboxItem,
  DropdownMenuRadioItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuShortcut,
  DropdownMenuGroup,
  DropdownMenuPortal,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
  DropdownMenuRadioGroup,
} from './Dropdown';

// Accordion
export { Accordion, AccordionItem, AccordionTrigger, AccordionContent } from './Accordion';

// Progress
export { Progress, CircularProgress } from './Progress';
export type { ProgressProps } from './Progress';

// Alert
export { Alert, AlertTitle, AlertDescription } from './Alert';

// Textarea
export { Textarea } from './Textarea';
export type { TextareaProps } from './Textarea';

// Separator
export { Separator } from './Separator';
